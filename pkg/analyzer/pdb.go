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
	"strings"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type PdbAnalyzer struct{}

func (PdbAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "PodDisruptionBudget"
	apiDoc := kubernetes.K8sApiReference{
		Kind: kind,
		ApiVersion: schema.GroupVersion{
			Group:   "policy",
			Version: "v1",
		},
		OpenapiSchema: a.OpenapiSchema,
	}

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	list, err := a.Client.GetClient().PolicyV1().PodDisruptionBudgets(a.Namespace).List(a.Context, a.ListOptions())
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, pdb := range list.Items {
		var failures []common.Failure

		// Before accessing the Conditions, check if they exist or not.
		if len(pdb.Status.Conditions) == 0 {
			continue
		}
		if pdb.Status.Conditions[0].Type == "DisruptionAllowed" && pdb.Status.Conditions[0].Status == "False" {
			var doc string
			if pdb.Spec.MaxUnavailable != nil {
				doc = apiDoc.GetApiDocV2("spec.maxUnavailable")
			}
			if pdb.Spec.MinAvailable != nil {
				doc = apiDoc.GetApiDocV2("spec.minAvailable")
			}
			// A null selector matches no pods, so such a PDB can never block an
			// eviction and is not worth reporting. Every other selector shape
			// does select pods, so a blocked PDB must be reported whether the
			// selector uses matchLabels, matchExpressions, or is empty.
			if pdb.Spec.Selector != nil {
				reason := pdb.Status.Conditions[0].Reason

				for k, v := range pdb.Spec.Selector.MatchLabels {
					failures = append(failures, common.Failure{
						Text:          fmt.Sprintf("%s, expected pdb pod label %s=%s", reason, k, v),
						KubernetesDoc: doc,
						Sensitive: []common.Sensitive{
							{
								Unmasked: k,
								Masked:   util.MaskString(k),
							},
							{
								Unmasked: v,
								Masked:   util.MaskString(v),
							},
						},
					})
				}

				for _, expr := range pdb.Spec.Selector.MatchExpressions {
					sensitive := []common.Sensitive{
						{
							Unmasked: expr.Key,
							Masked:   util.MaskString(expr.Key),
						},
					}
					for _, v := range expr.Values {
						sensitive = append(sensitive, common.Sensitive{
							Unmasked: v,
							Masked:   util.MaskString(v),
						})
					}

					text := fmt.Sprintf("%s, expected pdb pod label %s %s", reason, expr.Key, strings.ToLower(string(expr.Operator)))
					if len(expr.Values) > 0 {
						text = fmt.Sprintf("%s (%s)", text, strings.Join(expr.Values, ", "))
					}

					failures = append(failures, common.Failure{
						Text:          text,
						KubernetesDoc: doc,
						Sensitive:     sensitive,
					})
				}

				// An empty selector selects every pod in the namespace, so it
				// has the widest blast radius of all and must not be dropped.
				if len(pdb.Spec.Selector.MatchLabels) == 0 && len(pdb.Spec.Selector.MatchExpressions) == 0 {
					failures = append(failures, common.Failure{
						Text:          fmt.Sprintf("%s, and its empty selector applies to every pod in namespace %s", reason, pdb.Namespace),
						KubernetesDoc: doc,
						Sensitive: []common.Sensitive{
							{
								Unmasked: pdb.Namespace,
								Masked:   util.MaskString(pdb.Namespace),
							},
						},
					})
				}
			}
		}

		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", pdb.Namespace, pdb.Name)] = common.PreAnalysis{
				PodDisruptionBudget: pdb,
				FailureDetails:      failures,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, pdb.Name, pdb.Namespace).Set(float64(len(failures)))
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, found := util.GetParent(a.Client, value.PodDisruptionBudget.ObjectMeta)
		if found {
			currentAnalysis.ParentObject = parent
		}
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, err
}
