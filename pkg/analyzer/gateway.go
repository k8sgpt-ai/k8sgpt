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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	gtwapi "sigs.k8s.io/gateway-api/apis/v1"
)

type GatewayAnalyzer struct{}

// Gateway analyser will analyse all different Kinds and search for missing object dependencies
func (GatewayAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "Gateway"
	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	gtwList := &gtwapi.GatewayList{}
	gc := &gtwapi.GatewayClass{}
	client := a.Client.CtrlClient
	err := gtwapi.AddToScheme(client.Scheme())
	if err != nil {
		return nil, err
	}

	labelSelector := util.LabelStrToSelector(a.LabelSelector)
	if err := client.List(a.Context, gtwList, &ctrl.ListOptions{LabelSelector: labelSelector}); err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}
	// Find all unhealthy gateway Classes

	for _, gtw := range gtwList.Items {
		var failures []common.Failure

		gtwName := gtw.GetName()
		gtwNamespace := gtw.GetNamespace()
		// Check if gatewayclass exists
		err := client.Get(a.Context, ctrl.ObjectKey{Namespace: gtwNamespace, Name: string(gtw.Spec.GatewayClassName)}, gc, &ctrl.GetOptions{})
		if errors.IsNotFound(err) {
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf(
					"Gateway uses the GatewayClass %s which does not exist.",
					gtw.Spec.GatewayClassName,
				),
				Sensitive: []common.Sensitive{
					{
						Unmasked: string(gtw.Spec.GatewayClassName),
						Masked:   util.MaskString(string(gtw.Spec.GatewayClassName)),
					},
				},
			})
		}

		// Inspect Accepted and Programmed by type. Listener/address status is out of scope.
		for _, cond := range gtw.Status.Conditions {
			if cond.Type != string(gtwapi.GatewayConditionAccepted) &&
				cond.Type != string(gtwapi.GatewayConditionProgrammed) {
				continue
			}
			if cond.Status == metav1.ConditionTrue {
				continue
			}
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf("Gateway '%s/%s' has condition %s=%s, reason %s: %s",
					gtwNamespace,
					gtwName,
					cond.Type,
					cond.Status,
					cond.Reason,
					cond.Message,
				),
				Sensitive: []common.Sensitive{
					{
						Unmasked: gtwNamespace,
						Masked:   util.MaskString(gtwNamespace),
					},
					{
						Unmasked: gtwName,
						Masked:   util.MaskString(gtwName),
					},
				},
			})
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", gtwNamespace, gtwName)] = common.PreAnalysis{
				Gateway:        gtw,
				FailureDetails: failures,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, gtwName, gtwNamespace).Set(float64(len(failures)))
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
