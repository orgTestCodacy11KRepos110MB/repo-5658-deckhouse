/*
Copyright 2023 Flant JSC

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
	"fmt"
	"os"
	"path"
	"sort"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/flant/addon-operator/pkg/module_manager/go_hook"
	"github.com/flant/addon-operator/sdk"
	"github.com/flant/shell-operator/pkg/kube/object_patch"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/pointer"

	"github.com/deckhouse/deckhouse/modules/005-external-module-manager/hooks/internal/apis/v1alpha1"
)

var _ = sdk.RegisterFunc(&go_hook.HookConfig{
	Queue: "/modules/external-module-source/apply-release",
	Kubernetes: []go_hook.KubernetesConfig{
		{
			Name:                         "releases",
			ApiVersion:                   "deckhouse.io/v1alpha1",
			Kind:                         "ExternalModuleRelease",
			ExecuteHookOnEvents:          pointer.Bool(true),
			ExecuteHookOnSynchronization: pointer.Bool(true),
			FilterFunc:                   filterRelease,
		},
	},
	// Schedule: []go_hook.ScheduleConfig{
	// 	{
	// 		Name:    "check_deckhouse_release",
	// 		Crontab: "13 3 * * *",
	// 	},
	// },
}, applyModuleRelease)

func applyModuleRelease(input *go_hook.HookInput) error {
	snap := input.Snapshots["releases"]

	externalModulesDir := os.Getenv("EXTERNAL_MODULES_DIR")

	moduleReleases := make(map[string][]enqueueRelease, 0)

	for _, sn := range snap {
		if sn == nil {
			continue
		}
		rel := sn.(enqueueRelease)
		moduleReleases[rel.Module] = append(moduleReleases[rel.Module], rel)

		if rel.Status == "" {
			rel.Status = v1alpha1.PhasePending
		}
	}

	for module, releases := range moduleReleases {
		sort.Sort(byReleaseVersion(releases))

		var prevReleaseIndex = -1
		for i, rel := range releases {
			if rel.Status == v1alpha1.PhasePending {
				fmt.Println("APPLYING", rel)
				prevReleaseIndex = i
				status := map[string]v1alpha1.ExternalModuleReleaseStatus{
					"status": {
						Phase:          v1alpha1.PhaseDeployed,
						TransitionTime: time.Now().UTC(),
						Message:        "",
					},
				}

				symlinkName := path.Join(externalModulesDir, "modules", module)
				modulePath := path.Join(externalModulesDir, module, "v"+rel.Version.String())
				err := enableModule(symlinkName, modulePath)
				if err != nil {
					input.LogEntry.Errorf("Module release failed: %v", err)
					continue
				}
				input.PatchCollector.MergePatch(status, "deckhouse.io/v1alpha1", "ExternalModuleRelease", "", rel.Name, object_patch.WithSubresource("/status"))
				continue
			}

			if i == prevReleaseIndex {
				status := map[string]v1alpha1.ExternalModuleReleaseStatus{
					"status": {
						Phase:          v1alpha1.PhaseOutdated,
						TransitionTime: time.Now().UTC(),
						Message:        "",
					},
				}
				prevRelease := releases[prevReleaseIndex]
				input.PatchCollector.MergePatch(status, "deckhouse.io/v1alpha1", "ExternalModuleRelease", "", prevRelease.Name, object_patch.WithSubresource("/status"))

				continue
			}
		}
	}

	return nil
}

func enableModule(symlinkPath, modulePath string) error {
	if _, err := os.Lstat(symlinkPath); err == nil {
		err = os.Remove(symlinkPath)
		if err != nil {
			return err
		}
	}

	return os.Symlink(modulePath, symlinkPath)
	// TODO: send signal for reload
}

func filterRelease(obj *unstructured.Unstructured) (go_hook.FilterResult, error) {
	var release v1alpha1.ExternalModuleRelease

	err := sdk.FromUnstructured(obj, &release)
	if err != nil {
		return nil, err
	}

	var releaseApproved bool
	if v, ok := release.Annotations["release.deckhouse.io/approved"]; ok {
		if v == "true" {
			releaseApproved = true
		}
	}

	return enqueueRelease{
		Name:     release.Name,
		Version:  release.Spec.Version,
		Module:   release.Spec.ModuleName,
		Status:   release.Status.Phase,
		Approved: releaseApproved,
	}, nil
}

type enqueueRelease struct {
	Name     string
	Version  *semver.Version
	Module   string
	Status   string
	Approved bool
}
type byReleaseVersion []enqueueRelease

func (dr byReleaseVersion) Len() int {
	return len(dr)
}
func (dr byReleaseVersion) Swap(i, j int) {
	dr[i], dr[j] = dr[j], dr[i]
}
func (dr byReleaseVersion) Less(i, j int) bool {
	return dr[i].Version.LessThan(dr[j].Version)
}
