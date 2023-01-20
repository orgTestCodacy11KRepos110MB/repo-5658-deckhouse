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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/deckhouse/deckhouse/testing/hooks"
)

var _ = FDescribe("Modules :: deckhouse :: hooks :: set module image value ::", func() {
	f := HookExecutionConfigInit(`
global:
  deckhouseVersion: "12345"
  modulesImages:
    registry:
      base: registry.deckhouse.io/deckhouse/fe
externalModuleSource:
  internal: {}
`, `{}`)
	f.RegisterCRD("deckhouse.io", "v1alpha1", "ExternalModuleSource", false)

	Context("With Deckhouse pod", func() {
		BeforeEach(func() {
			sc := f.KubeStateSet(`
apiVersion: deckhouse.io/v1alpha1
kind: ExternalModuleSource
metadata:
  name: registry-deckhouse
spec:
  registry:
    repo: dev-registry.deckhouse.io/deckhouse/external-modules
    dockerCfg: "d3JpdGV0b2tlbjo0YWM5ZmUyYjcwYzU0ZDA5ZjZkNDM1MzI4ZGZmMDg0NA=="
`)

			// EXTERNAL_MODULE_DIRECTORY
			// MODULES_DIR
			// /external-modules    -----> hostPath
			//   /echoserver/v0.0.1
			//   /echoserver/v0.0.2
			//   /foobar/v1.0.0
			//   /modules
			//     @/900-echoserver -> ../echoserver/v0.0.2
			//     @/901-foobar -> ../foobar/v1.0.0
			//     /903-dev-module

			// ❯ crane ls dev-registry.deckhouse.io/deckhouse/external-modules
			// echoserver
			// ❯ crane ls dev-registry.deckhouse.io/deckhouse/external-modules/echoserver/release
			// beta
			// v0.0.1
			// v1.0.0
			// example
			// ❯ crane pull dev-registry.deckhouse.io/deckhouse/external-modules/echoserver/release:beta 111.tar
			// ❯ tar -xf 111.tar
			// ❯ tar -xf 8461a4b028e7a9c3caa9f34bf988a33b30f9b0415473d6efb616c23855c1acab.tar.gz
			// ❯ cat version.json
			// {"version":"v0.0.1"}

			// dev-registry.deckhouse.io/deckhouse/external-modules/echoserver:v0.0.1

			f.BindingContexts.Set(sc)
			f.RunHook()
		})

		It("boo", func() {
			Expect(f).To(ExecuteSuccessfully())
			res := f.KubernetesGlobalResource("ExternalModuleSource", "registry-deckhouse")
			fmt.Println(res.ToYaml())
		})
	})
})
