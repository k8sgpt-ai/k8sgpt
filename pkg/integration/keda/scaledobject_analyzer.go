package keda

import (
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	kedaSchema "github.com/kedacore/keda/v2/apis/keda/v1alpha1"
	"github.com/kedacore/keda/v2/pkg/generated/clientset/versioned/typed/keda/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ScaledObjectAnalyzer struct{}

func (s *ScaledObjectAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	kClient, _ := v1alpha1.NewForConfig(a.Client.GetConfig())
	kind := "ScaledObject"

	apiDoc := kubernetes.K8sApiReference{
		Kind:          kind,
		ApiVersion:    kedaSchema.GroupVersion,
		OpenapiSchema: a.OpenapiSchema,
	}

	list, err := kClient.ScaledObjects(a.Namespace).List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, so := range list.Items {
		var failures []common.Failure

		scaleTargetRef := so.Spec.ScaleTargetRef
		if scaleTargetRef.Kind == "" {
			scaleTargetRef.Kind = "Deployment"
		}

		var podInfo PodInfo

		switch scaleTargetRef.Kind {
		case "Deployment":
			deployment, err := a.Client.GetClient().AppsV1().Deployments(so.Namespace).Get(a.Context, scaleTargetRef.Name, metav1.GetOptions{})
			if err == nil {
				podInfo = DeploymentInfo{deployment}
			}
		case "ReplicationController":
			rc, err := a.Client.GetClient().CoreV1().ReplicationControllers(so.Namespace).Get(a.Context, scaleTargetRef.Name, metav1.GetOptions{})
			if err == nil {
				podInfo = ReplicationControllerInfo{rc}
			}
		case "ReplicaSet":
			rs, err := a.Client.GetClient().AppsV1().ReplicaSets(so.Namespace).Get(a.Context, scaleTargetRef.Name, metav1.GetOptions{})
			if err == nil {
				podInfo = ReplicaSetInfo{rs}
			}
		case "StatefulSet":
			ss, err := a.Client.GetClient().AppsV1().StatefulSets(so.Namespace).Get(a.Context, scaleTargetRef.Name, metav1.GetOptions{})
			if err == nil {
				podInfo = StatefulSetInfo{ss}
			}
		default:
			failures = append(failures, common.Failure{
				Text:      fmt.Sprintf("ScaledObject uses %s as ScaleTargetRef which is not an option.", scaleTargetRef.Kind),
				Sensitive: []common.Sensitive{},
			})
		}

		if podInfo == nil {
			doc := apiDoc.GetApiDocV2("spec.scaleTargetRef")

			failures = append(failures, common.Failure{
				Text:          fmt.Sprintf("ScaledObject uses %s/%s as ScaleTargetRef which does not exist.", scaleTargetRef.Kind, scaleTargetRef.Name),
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
				for _, trigger := range so.Spec.Triggers {
					if trigger.Type == "cpu" || trigger.Type == "memory" {
						if container.Resources.Requests == nil || container.Resources.Limits == nil {
							containers--
							break
						}
					}
				}
			}

			if containers <= 0 {
				doc := apiDoc.GetApiDocV2("spec.scaleTargetRef.kind")

				failures = append(failures, common.Failure{
					Text:          fmt.Sprintf("%s %s/%s does not have resource configured.", scaleTargetRef.Kind, so.Namespace, scaleTargetRef.Name),
					KubernetesDoc: doc,
					Sensitive: []common.Sensitive{
						{
							Unmasked: scaleTargetRef.Name,
							Masked:   util.MaskString(scaleTargetRef.Name),
						},
					},
				})
			}

			evt, err := util.FetchLatestEvent(a.Context, a.Client, so.Namespace, so.Name)
			if err != nil || evt == nil {
				continue
			}

			if evt.Type != "Normal" {
				failures = append(failures, common.Failure{
					Text: evt.Message,
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
			preAnalysis[fmt.Sprintf("%s/%s", so.Namespace, so.Name)] = common.PreAnalysis{
				ScaledObject:   so,
				FailureDetails: failures,
			}
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.ScaledObject.ObjectMeta)
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
