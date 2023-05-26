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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type HpaAnalyzer struct{}

func (HpaAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "HorizontalPodAutoscaler"
	apiDoc := kubernetes.K8sApiReference{
		Kind: kind,
		ApiVersion: schema.GroupVersion{
			Group:   "autoscaling",
			Version: "v1",
		},
		Discovery: a.Client.Client.Discovery(),
	}

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	list, err := a.Client.GetClient().AutoscalingV1().HorizontalPodAutoscalers(a.Namespace).List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, hpa := range list.Items {
		var failures []common.Failure

		// check ScaleTargetRef exist
		scaleTargetRef := hpa.Spec.ScaleTargetRef
		var podInfo PodInfo

		switch scaleTargetRef.Kind {
		case "Deployment":
			deployment, err := a.Client.GetClient().AppsV1().Deployments(hpa.Namespace).Get(a.Context, scaleTargetRef.Name, metav1.GetOptions{})
			if err == nil {
				podInfo = DeploymentInfo{deployment}
			}
		case "ReplicationController":
			rc, err := a.Client.GetClient().CoreV1().ReplicationControllers(hpa.Namespace).Get(a.Context, scaleTargetRef.Name, metav1.GetOptions{})
			if err == nil {
				podInfo = ReplicationControllerInfo{rc}
			}
		case "ReplicaSet":
			rs, err := a.Client.GetClient().AppsV1().ReplicaSets(hpa.Namespace).Get(a.Context, scaleTargetRef.Name, metav1.GetOptions{})
			if err == nil {
				podInfo = ReplicaSetInfo{rs}
			}
		case "StatefulSet":
			ss, err := a.Client.GetClient().AppsV1().StatefulSets(hpa.Namespace).Get(a.Context, scaleTargetRef.Name, metav1.GetOptions{})
			if err == nil {
				podInfo = StatefulSetInfo{ss}
			}
		default:
			failures = append(failures, common.Failure{
				Text:      fmt.Sprintf("HorizontalPodAutoscaler uses %s as ScaleTargetRef which is not an option.", scaleTargetRef.Kind),
				Sensitive: []common.Sensitive{},
			})
		}

		if podInfo == nil {
			doc := apiDoc.GetApiDocV2("spec.scaleTargetRef")

			failures = append(failures, common.Failure{
				Text:          fmt.Sprintf("HorizontalPodAutoscaler uses %s/%s as ScaleTargetRef which does not exist.", scaleTargetRef.Kind, scaleTargetRef.Name),
				KubernetesDoc: doc,
				Sensitive: []common.Sensitive{
					{
						Unmasked: scaleTargetRef.Name,
						Masked:   util.MaskString(scaleTargetRef.Name),
					},
				},
			})
		} else {
			containers := len(podInfo.GetPodSpec().Containers)
			for _, container := range podInfo.GetPodSpec().Containers {
				if container.Resources.Requests == nil || container.Resources.Limits == nil {
					containers--
				}
			}

			if containers <= 0 {
				doc := apiDoc.GetApiDocV2("spec.scaleTargetRef.kind")

				failures = append(failures, common.Failure{
					Text:          fmt.Sprintf("%s %s/%s does not have resource configured.", scaleTargetRef.Kind, a.Namespace, scaleTargetRef.Name),
					KubernetesDoc: doc,
					Sensitive: []common.Sensitive{
						{
							Unmasked: scaleTargetRef.Name,
							Masked:   util.MaskString(scaleTargetRef.Name),
						},
					},
				})
			}

		}

		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", hpa.Namespace, hpa.Name)] = common.PreAnalysis{
				HorizontalPodAutoscalers: hpa,
				FailureDetails:           failures,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, hpa.Name, hpa.Namespace).Set(float64(len(failures)))
		}

	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.HorizontalPodAutoscalers.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}

type PodInfo interface {
	GetPodSpec() corev1.PodSpec
}

type DeploymentInfo struct {
	*appsv1.Deployment
}

func (d DeploymentInfo) GetPodSpec() corev1.PodSpec {
	return d.Spec.Template.Spec
}

// define a structure for ReplicationController
type ReplicationControllerInfo struct {
	*corev1.ReplicationController
}

func (rc ReplicationControllerInfo) GetPodSpec() corev1.PodSpec {
	return rc.Spec.Template.Spec
}

// define a structure for ReplicaSet
type ReplicaSetInfo struct {
	*appsv1.ReplicaSet
}

func (rs ReplicaSetInfo) GetPodSpec() corev1.PodSpec {
	return rs.Spec.Template.Spec
}

// define a structure for StatefulSet
type StatefulSetInfo struct {
	*appsv1.StatefulSet
}

// implement PodInfo for StatefulSetInfo
func (ss StatefulSetInfo) GetPodSpec() corev1.PodSpec {
	return ss.Spec.Template.Spec
}
