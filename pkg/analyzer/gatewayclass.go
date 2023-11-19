/*
Copyright 2023 The K8sGPT Authors.
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
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	gtwapi "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	AcceptedStatus = "True"
	gtwcResource   = "gatewayclasses"
)

type GatewayClassAnalyzer struct{}

// Gateway analyser will analyse all different Kinds and search for missing object dependencies
func (GatewayClassAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "GatewayClass"
	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	gtwcList := &gtwapi.GatewayClassList{}
	client := a.Client.DynClient
	runtimeObj, _ := client.Resource(gtwapi.SchemeGroupVersion.WithResource(gtwcResource)).List((a.Context), metav1.ListOptions{})
	// converting to schema
	runtime.DefaultUnstructuredConverter.FromUnstructured(runtimeObj.UnstructuredContent(), gtwcList)
	var preAnalysis = map[string]common.PreAnalysis{}

	// Find all unhealthy gateway Classes

	for _, gc := range gtwcList.Items {
		var failures []common.Failure

		gcName := gc.GetName()
		// Check only the current condition
		if gc.Status.Conditions[0].Status != AcceptedStatus {
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf(
					"GatewayClass '%s' with a controller name '%s' is not accepted. Message: '%s'.",
					gcName,
					gc.Spec.ControllerName,
					gc.Status.Conditions[0].Message,
				),
				Sensitive: []common.Sensitive{
					{
						Unmasked: gcName,
						Masked:   util.MaskString(gcName),
					},
				},
			})
		}
		if len(failures) > 0 {
			preAnalysis[gcName] = common.PreAnalysis{
				GatewayClass:   gc,
				FailureDetails: failures,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, gcName, "").Set(float64(len(failures)))
		}
	}
	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}
		a.Results = append(a.Results, currentAnalysis)
	}
	return a.Results, nil
}
