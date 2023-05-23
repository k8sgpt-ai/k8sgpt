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
)

type StatefulSetAnalyzer struct{}

func (StatefulSetAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "StatefulSet"
	apiDoc := util.K8sApiReference{
		Kind:          kind,
		ApiVersion:    "apps/v1",
		ServerVersion: a.Client.ServerVersion,
	}

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	list, err := a.Client.GetClient().AppsV1().StatefulSets(a.Namespace).List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var preAnalysis = map[string]common.PreAnalysis{}

	for _, sts := range list.Items {
		var failures []common.Failure

		// get serviceName
		serviceName := sts.Spec.ServiceName
		_, err := a.Client.GetClient().CoreV1().Services(sts.Namespace).Get(a.Context, serviceName, metav1.GetOptions{})
		if err != nil {
			doc, _ := apiDoc.GetApiDoc("serviceName")
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf(
					"StatefulSet uses the service %s/%s which does not exist.\n  Official Doc: %s",
					sts.Namespace,
					serviceName,
					doc,
				),
				Sensitive: []common.Sensitive{
					{
						Unmasked: sts.Namespace,
						Masked:   util.MaskString(sts.Namespace),
					},
					{
						Unmasked: serviceName,
						Masked:   util.MaskString(serviceName),
					},
				},
			})
		}
		if len(sts.Spec.VolumeClaimTemplates) > 0 {
			for _, volumeClaimTemplate := range sts.Spec.VolumeClaimTemplates {
				if volumeClaimTemplate.Spec.StorageClassName != nil {
					_, err := a.Client.GetClient().StorageV1().StorageClasses().Get(a.Context, *volumeClaimTemplate.Spec.StorageClassName, metav1.GetOptions{})
					if err != nil {
						failures = append(failures, common.Failure{
							Text: fmt.Sprintf("StatefulSet uses the storage class %s which does not exist.", *volumeClaimTemplate.Spec.StorageClassName),
							Sensitive: []common.Sensitive{
								{
									Unmasked: *volumeClaimTemplate.Spec.StorageClassName,
									Masked:   util.MaskString(*volumeClaimTemplate.Spec.StorageClassName),
								},
							},
						})
					}
				}
			}
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", sts.Namespace, sts.Name)] = common.PreAnalysis{
				StatefulSet:    sts,
				FailureDetails: failures,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, sts.Name, sts.Namespace).Set(float64(len(failures)))
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.StatefulSet.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}
