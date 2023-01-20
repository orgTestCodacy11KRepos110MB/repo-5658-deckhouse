/*
Copyright 2022 Flant JSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hooks

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/flant/addon-operator/pkg/module_manager/go_hook"
	"github.com/flant/addon-operator/sdk"
	"github.com/flant/shell-operator/pkg/kube/object_patch"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/pointer"

	"github.com/deckhouse/deckhouse/go_lib/dependency/cr"

	"github.com/deckhouse/deckhouse/go_lib/dependency"
)

var _ = sdk.RegisterFunc(&go_hook.HookConfig{
	Queue: "/modules/external-module-source/sources",
	Kubernetes: []go_hook.KubernetesConfig{
		{
			Name:                "sources",
			ApiVersion:          "deckhouse.io/v1alpha1",
			Kind:                "ExternalModuleSource",
			ExecuteHookOnEvents: pointer.Bool(true),
			FilterFunc:          filterSource,
		},
	},
	// Schedule: []go_hook.ScheduleConfig{
	// 	{
	// 		Name:    "check_deckhouse_release",
	// 		Crontab: "* * * * *", // every minute
	// 	},
	// },
}, dependency.WithExternalDependencies(xxx))

func filterSource(obj *unstructured.Unstructured) (go_hook.FilterResult, error) {
	var ex ExternalModuleSource

	err := sdk.FromUnstructured(obj, &ex)

	return ex, err
}

func xxx(input *go_hook.HookInput, dc dependency.Container) error {
	ts := time.Now().UTC()
	snap := input.Snapshots["sources"]

	fmt.Println("SNAP", len(snap))

	for _, sn := range snap {
		ex := sn.(ExternalModuleSource)
		sc := ExternalModuleSourceStatus{
			SyncTime: ts,
		}

		opts := make([]cr.Option, 0)
		if ex.Spec.Registry.DockerCFG != "" {
			opts = append(opts, cr.WithAuth(ex.Spec.Registry.DockerCFG))
		} else {
			opts = append(opts, cr.WithDisabledAuth())
		}

		regCli, err := dc.GetRegistryClient(ex.Spec.Registry.Repo, opts...)
		if err != nil {
			sc.Msg = err.Error()
			updateSourceStatus(input, ex.Name, sc)
			continue
		}

		tags, err := regCli.ListTags()
		if err != nil {
			sc.Msg = err.Error()
			updateSourceStatus(input, ex.Name, sc)
			continue
		}

		sort.Strings(tags)

		sc.Msg = ""
		sc.AvailableModules = tags
		updateSourceStatus(input, ex.Name, sc)

		// fetch release image
		// TODO: get release channel from somewhere
		const releaseChannel = "beta"

		// TODO: check digest
		// dev-registry.deckhouse.io/deckhouse/external-modules/echoserver/release
		for _, moduleName := range tags {
			regCli, _ = dc.GetRegistryClient(path.Join(ex.Spec.Registry.Repo, moduleName, "release"), opts...)
			// TODO: save digest
			img, err := regCli.Image(releaseChannel)
			if err != nil {
				// TODO: err
				fmt.Println("ERROR2", err)
				continue
			}

			digest, err := img.Digest()
			if err != nil {
				// TODO: err
				fmt.Println("ERROR22", err)
				continue
			}
			fmt.Println("DIGEST", moduleName, releaseChannel, digest)

			meta, err := fetchModuleReleaseMetadata(img)
			if err != nil {
				fmt.Println("ERROR223", err)
				continue
			}

			fmt.Println("GOT MODULE METADATA", meta)
			regCli, _ = dc.GetRegistryClient(path.Join(ex.Spec.Registry.Repo, moduleName), opts...)

			img, err = regCli.Image(meta.Version.Original())
			if err != nil {
				return err
			}
			err = copyModuleToFS("/tmp/bbbc", img)
			if err != nil {
				return err
			}
		}
		// end fetch release
		fmt.Println("COPIED")
	}

	return nil
}

func copyModuleToFS(rootPath string, img v1.Image) error {
	layers, err := img.Layers()
	if err != nil {
		return err
	}

	for _, layer := range layers {
		rc, err := layer.Uncompressed()
		if err != nil {
			return err
		}
		err = copyLayerToFS(rootPath, rc)
		if err != nil {
			return err
		}
	}

	return nil
}

func copyLayerToFS(rootPath string, rc io.ReadCloser) error {
	tr := tar.NewReader(rc)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of archive
			return nil
		}
		if err != nil {
			return err
		}

		if strings.HasSuffix(hdr.Name, ".wh..wh..opq") {
			continue
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path.Join(rootPath, hdr.Name), 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(path.Join(rootPath, hdr.Name))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()

			err = os.Chmod(outFile.Name(), os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}

		default:
			return errors.New("unknown tar type")
		}
	}
}

func untarVersionLayer(rc io.ReadCloser, rw io.Writer) error {
	tr := tar.NewReader(rc)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of archive
			return nil
		}
		if err != nil {
			return err
		}
		if strings.HasPrefix(hdr.Name, ".werf") {
			continue
		}

		switch hdr.Name {
		case "version.json":
			_, err = io.Copy(rw, tr)
			if err != nil {
				return err
			}
			return nil

		default:
			continue
		}
	}
}

func fetchModuleReleaseMetadata(img v1.Image) (moduleReleaseMetadata, error) {
	buf := bytes.NewBuffer(nil)
	var meta moduleReleaseMetadata

	layers, err := img.Layers()
	if err != nil {
		return meta, err
	}

	for _, layer := range layers {
		size, err := layer.Size()
		if err != nil {
			// dcr.logger.Warnf("couldn't calculate layer size")
			return meta, err
		}
		if size == 0 {
			// skip some empty werf layers
			continue
		}
		rc, err := layer.Uncompressed()
		if err != nil {
			return meta, err
		}

		err = untarVersionLayer(rc, buf)
		if err != nil {
			return meta, err
		}

		rc.Close()
	}

	err = json.Unmarshal(buf.Bytes(), &meta)

	return meta, err
}

type moduleReleaseMetadata struct {
	Version semver.Version `json:"version"`
}

func updateSourceStatus(input *go_hook.HookInput, name string, sc ExternalModuleSourceStatus) {
	st := map[string]ExternalModuleSourceStatus{
		"status": sc,
	}

	input.PatchCollector.MergePatch(st, "deckhouse.io/v1alpha1", "ExternalModuleSource", "", name, object_patch.WithSubresource("/status"))
}

type ExternalModuleSource struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of an ExternalModuleSource.
	Spec ExternalModuleSourceSpec `json:"spec"`

	// Status of an ExternalModuleSource.
	Status ExternalModuleSourceStatus `json:"status,omitempty"`
}

type ExternalModuleSourceSpec struct {
	Registry struct {
		Repo      string `json:"repo"`
		DockerCFG string `json:"dockerCfg"`
	} `json:"registry"`
}

type ExternalModuleSourceStatus struct {
	SyncTime         time.Time `json:"syncTime,omitempty"`
	AvailableModules []string  `json:"availableModules,omitempty"`
	Msg              string    `json:"message,omitempty"`
}
