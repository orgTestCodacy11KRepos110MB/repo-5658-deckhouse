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
	"github.com/deckhouse/deckhouse/go_lib/hooks/tls_certificate"
)

var _ = tls_certificate.RegisterInternalTLSHook(tls_certificate.GenSelfSignedTLSHookConf{
	SANs: tls_certificate.DefaultSANs([]string{
		"gatekeeper-webhook-service.d8-admission-policy-engine.svc",
		tls_certificate.ClusterDomainSAN("gatekeeper-webhook-service.d8-admission-policy-engine.svc"),
	}),

	CN: "gatekeeper-webhook-service.d8-admission-policy-engine.svc",

	Namespace:            "d8-admission-policy-engine",
	TLSSecretName:        "gatekeeper-webhook-server-cert",
	FullValuesPathPrefix: "admissionPolicyEngine.internal.webhook",
})
