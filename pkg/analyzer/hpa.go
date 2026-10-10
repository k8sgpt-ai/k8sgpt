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
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// scalingLimitedTooFewReplicas is the reason the HPA controller sets on the
// ScalingLimited condition when the raw replica calculation is below
// spec.minReplicas. Kubernetes does not export these reason strings, so the
// value is mirrored here.
const scalingLimitedTooFewReplicas = "TooFewReplicas"

type HpaAnalyzer struct{}

func (HpaAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "HorizontalPodAutoscaler"
	apiDoc := kubernetes.K8sApiReference{
		Kind: kind,
		ApiVersion: schema.GroupVersion{
			Group:   "autoscaling",
			Version: "v2",
		},
		OpenapiSchema: a.OpenapiSchema,
	}

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	list, err := a.Client.GetClient().AutoscalingV2().HorizontalPodAutoscalers(a.Namespace).List(a.Context, a.ListOptions())
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, hpa := range list.Items {
		var failures []common.Failure

		//check the error from status field
		conditions := hpa.Status.Conditions
		for _, condition := range conditions {
			// https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/#appendix-horizontal-pod-autoscaler-status-conditions
			switch condition.Type {
			case autoscalingv2.ScalingLimited:
				if condition.Status != corev1.ConditionTrue {
					break
				}
				// An HPA sitting at its minimum replica count always reports
				// ScalingLimited=True with reason TooFewReplicas. That is the
				// replica floor working as designed, not a fault.
				atMinReplicas := hpa.Spec.MinReplicas != nil &&
					hpa.Status.DesiredReplicas == *hpa.Spec.MinReplicas
				if condition.Reason == scalingLimitedTooFewReplicas && atMinReplicas {
					break
				}
				failures = append(failures, common.Failure{
					Text:      condition.Message,
					Sensitive: []common.Sensitive{},
				})
			default:
				if condition.Status == corev1.ConditionFalse {
					failures = append(failures, common.Failure{
						Text:      condition.Message,
						Sensitive: []common.Sensitive{},
					})
				}
			}
		}

		// check ScaleTargetRef exist
		scaleTargetRef := hpa.Spec.ScaleTargetRef
		var podInfo PodInfo
		supportedKind := true

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
			supportedKind = false
			failures = append(failures, common.Failure{
				Text:      fmt.Sprintf("HorizontalPodAutoscaler uses %s as ScaleTargetRef which is not an option.", scaleTargetRef.Kind),
				Sensitive: []common.Sensitive{},
			})
		}

		if podInfo == nil {
			// an unsupported kind was never looked up, so it is not a missing
			// target and the failure for it has already been recorded
			if supportedKind {
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
			}
		} else {
			podSpec := podInfo.GetPodSpec()
			for _, target := range utilizationTargets(hpa) {
				missing, found := containersMissingRequest(podSpec, target)

				var text string
				switch {
				case !found:
					text = fmt.Sprintf("HorizontalPodAutoscaler scales on the %s utilization of container %s, which does not exist in %s %s/%s.", target.resource, target.container, scaleTargetRef.Kind, hpa.Namespace, scaleTargetRef.Name)
				case len(missing) > 0:
					text = fmt.Sprintf("%s %s/%s does not set a %s request on container(s) %s, so the HorizontalPodAutoscaler cannot compute its %s utilization.", scaleTargetRef.Kind, hpa.Namespace, scaleTargetRef.Name, target.resource, strings.Join(missing, ", "), target.resource)
				default:
					continue
				}

				failures = append(failures, common.Failure{
					Text:          text,
					KubernetesDoc: apiDoc.GetApiDocV2("spec.metrics"),
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

		parent, found := util.GetParent(a.Client, value.HorizontalPodAutoscalers.ObjectMeta)
		if found {
			currentAnalysis.ParentObject = parent
		}
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}

// hpaUtilizationTarget is a resource whose utilization the HPA controller
// computes as a percentage of the scale target's resource requests. An empty
// container means the requests of the whole pod are used.
type hpaUtilizationTarget struct {
	resource  corev1.ResourceName
	container string
}

// utilizationTargets returns the Resource and ContainerResource metrics of the
// HPA that are measured as utilization. Only those depend on resource
// requests: an AverageValue target compares raw usage, and Pods, Object and
// External metrics never read requests. An HPA without metrics is defaulted by
// the API server to 80% average CPU utilization, so it is treated the same.
func utilizationTargets(hpa autoscalingv2.HorizontalPodAutoscaler) []hpaUtilizationTarget {
	if len(hpa.Spec.Metrics) == 0 {
		return []hpaUtilizationTarget{{resource: corev1.ResourceCPU}}
	}

	var targets []hpaUtilizationTarget
	seen := map[hpaUtilizationTarget]bool{}
	for _, metric := range hpa.Spec.Metrics {
		var target hpaUtilizationTarget
		switch {
		case metric.Type == autoscalingv2.ResourceMetricSourceType && metric.Resource != nil &&
			isUtilizationTarget(metric.Resource.Target):
			target = hpaUtilizationTarget{resource: metric.Resource.Name}
		case metric.Type == autoscalingv2.ContainerResourceMetricSourceType && metric.ContainerResource != nil &&
			isUtilizationTarget(metric.ContainerResource.Target):
			target = hpaUtilizationTarget{resource: metric.ContainerResource.Name, container: metric.ContainerResource.Container}
		default:
			continue
		}

		if !seen[target] {
			seen[target] = true
			targets = append(targets, target)
		}
	}
	return targets
}

// isUtilizationTarget mirrors the HPA controller, which uses an AverageValue
// target whenever one is set and only otherwise falls back to
// AverageUtilization.
func isUtilizationTarget(target autoscalingv2.MetricTarget) bool {
	return target.AverageValue == nil && target.AverageUtilization != nil
}

// containersMissingRequest returns the containers the HPA controller reads a
// request for target.resource from that do not set one, and whether the
// container named by target exists in the pod.
//
// The controller reads the requests of the created pods, not of the pod
// template, so the pod defaulting is taken into account: a resource with a
// limit but no request gets a request equal to the limit. When no container is
// named and the pod sets pod-level resources, the controller uses the
// pod-level request, which also falls back to the requests of the containers,
// so any of them is enough. Otherwise every app and sidecar container, or only
// the named one, needs a request.
func containersMissingRequest(podSpec corev1.PodSpec, target hpaUtilizationTarget) (missing []string, found bool) {
	if target.container == "" && setsPodLevelResources(podSpec) {
		if setsRequest(*podSpec.Resources, target.resource) {
			return nil, true
		}
		for _, containers := range [][]corev1.Container{podSpec.InitContainers, podSpec.Containers} {
			for _, c := range containers {
				if setsRequest(c.Resources, target.resource) {
					return nil, true
				}
			}
		}
	}

	containers := append([]corev1.Container{}, podSpec.Containers...)
	for _, c := range podSpec.InitContainers {
		if c.RestartPolicy != nil && *c.RestartPolicy == corev1.ContainerRestartPolicyAlways {
			containers = append(containers, c)
		}
	}

	found = target.container == ""
	for _, c := range containers {
		if target.container != "" && target.container != c.Name {
			continue
		}
		found = true
		if !setsRequest(c.Resources, target.resource) {
			missing = append(missing, c.Name)
		}
	}
	return missing, found
}

// setsRequest reports whether a pod created from resources has a request for
// name. The API server defaults a missing request to the limit on pods, but
// not on pod templates, so a limit alone is enough.
func setsRequest(resources corev1.ResourceRequirements, name corev1.ResourceName) bool {
	_, hasRequest := resources.Requests[name]
	_, hasLimit := resources.Limits[name]
	return hasRequest || hasLimit
}

// setsPodLevelResources reports whether a pod created from podSpec has
// pod-level requests, which the HPA controller then uses instead of summing
// the requests of the containers. Only CPU, memory and hugepages are supported
// at the pod level, and a pod-level limit defaults the pod-level request.
func setsPodLevelResources(podSpec corev1.PodSpec) bool {
	if podSpec.Resources == nil {
		return false
	}
	for _, list := range []corev1.ResourceList{podSpec.Resources.Requests, podSpec.Resources.Limits} {
		for name := range list {
			if name == corev1.ResourceCPU || name == corev1.ResourceMemory ||
				strings.HasPrefix(string(name), corev1.ResourceHugePagesPrefix) {
				return true
			}
		}
	}
	return false
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
