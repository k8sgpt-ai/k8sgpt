/*
Copyright 2026 The K8sGPT Authors.
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

package analyzer

import (
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ResourceClaimAnalyzer reports ResourceClaims that reference missing
// DeviceClasses and therefore cannot be allocated by the DRA scheduler.
type ResourceClaimAnalyzer struct{}

func (ResourceClaimAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	const kind = "ResourceClaim"

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	apiDoc := kubernetes.K8sApiReference{
		Kind: kind,
		ApiVersion: schema.GroupVersion{
			Group:   "resource.k8s.io",
			Version: "v1beta1",
		},
		OpenapiSchema: a.OpenapiSchema,
	}

	client := a.Client.GetClient().ResourceV1beta1()

	// DeviceClass is cluster-scoped. Do not apply a.LabelSelector here: a
	// selector limits which claims are analyzed, but must not hide a referenced
	// DeviceClass and turn it into a false missing-reference finding.
	deviceClasses, err := client.DeviceClasses().List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	knownDeviceClasses := make(map[string]struct{}, len(deviceClasses.Items))
	for _, deviceClass := range deviceClasses.Items {
		knownDeviceClasses[deviceClass.Name] = struct{}{}
	}

	claims, err := client.ResourceClaims(a.Namespace).List(a.Context, metav1.ListOptions{
		LabelSelector: a.LabelSelector,
	})
	if err != nil {
		return nil, err
	}

	var results []common.Result
	for _, claim := range claims.Items {
		var failures []common.Failure
		reportedDeviceClasses := make(map[string]struct{})

		for _, request := range claim.Spec.Devices.Requests {
			if request.DeviceClassName == "" {
				continue
			}
			if _, ok := knownDeviceClasses[request.DeviceClassName]; ok {
				continue
			}
			if _, ok := reportedDeviceClasses[request.DeviceClassName]; ok {
				continue
			}
			reportedDeviceClasses[request.DeviceClassName] = struct{}{}

			failures = append(failures, common.Failure{
				Text: fmt.Sprintf(
					"ResourceClaim %s references DeviceClass %s which does not exist.",
					claim.Name,
					request.DeviceClassName,
				),
				KubernetesDoc: apiDoc.GetApiDocV2("spec.devices.requests.deviceClassName"),
				Sensitive: []common.Sensitive{
					{
						Unmasked: claim.Name,
						Masked:   util.MaskString(claim.Name),
					},
					{
						Unmasked: request.DeviceClassName,
						Masked:   util.MaskString(request.DeviceClassName),
					},
				},
			})
		}

		if len(failures) == 0 {
			continue
		}

		name := fmt.Sprintf("%s/%s", claim.Namespace, claim.Name)
		results = append(results, common.Result{
			Kind:  kind,
			Name:  name,
			Error: failures,
		})
		AnalyzerErrorsMetric.WithLabelValues(kind, claim.Name, claim.Namespace).Set(float64(len(failures)))
	}

	return results, nil
}
