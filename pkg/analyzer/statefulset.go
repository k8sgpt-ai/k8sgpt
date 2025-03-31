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
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type StatefulSetAnalyzer struct{}

func (StatefulSetAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "StatefulSet"
	apiDoc := kubernetes.K8sApiReference{
		Kind: kind,
		ApiVersion: schema.GroupVersion{
			Group:   "apps",
			Version: "v1",
		},
		OpenapiSchema: a.OpenapiSchema,
	}

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	list, err := a.Client.GetClient().AppsV1().StatefulSets(a.Namespace).List(a.Context, metav1.ListOptions{LabelSelector: a.LabelSelector})
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
			doc := apiDoc.GetApiDocV2("spec.serviceName")

			failures = append(failures, common.Failure{
				Text: fmt.Sprintf(
					"StatefulSet uses the service %s/%s which does not exist.",
					sts.Namespace,
					serviceName,
				),
				KubernetesDoc: doc,
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
		if sts.Spec.Replicas != nil && *(sts.Spec.Replicas) != sts.Status.AvailableReplicas {
			for i := int32(0); i < *(sts.Spec.Replicas); i++ {
				podName := sts.Name + "-" + fmt.Sprint(i)
				pod, err := a.Client.GetClient().CoreV1().Pods(sts.Namespace).Get(a.Context, podName, metav1.GetOptions{})
				if err != nil {
					if errors.IsNotFound(err) && i == 0 {
						evt, err := util.FetchLatestEvent(a.Context, a.Client, sts.Namespace, sts.Name)
						if err != nil || evt == nil || evt.Type == "Normal" {
							break
						}
						failures = append(failures, common.Failure{
							Text:      evt.Message,
							Sensitive: []common.Sensitive{},
						})
					}
					break
				}
				if pod.Status.Phase != "Running" {
					failures = append(failures, common.Failure{
						Text: fmt.Sprintf("Statefulset pod %s in the namespace %s is not in running state.", pod.Name, pod.Namespace),
						Sensitive: []common.Sensitive{
							{
								Unmasked: sts.Namespace,
								Masked:   util.MaskString(pod.Name),
							},
							{
								Unmasked: serviceName,
								Masked:   util.MaskString(pod.Namespace),
							},
						},
					})
					break
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

		parent, found := util.GetParent(a.Client, value.StatefulSet.ObjectMeta)
		if found {
			currentAnalysis.ParentObject = parent
		}
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}
