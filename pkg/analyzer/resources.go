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
	"strings"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourcesAnalyzer checks for missing CPU/memory requests/limits on workload containers.
//
// This intentionally starts with Deployments only (safe, small surface) and can be extended
// to other workload kinds in follow-up PRs.
type ResourcesAnalyzer struct{}

func (ResourcesAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	kind := "Resources"

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	deployments, err := a.Client.GetClient().AppsV1().Deployments(a.Namespace).List(a.Context, metav1.ListOptions{
		LabelSelector: a.LabelSelector,
	})
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, deployment := range deployments.Items {
		var failures []common.Failure

		containers := deployment.Spec.Template.Spec.Containers
		initContainers := deployment.Spec.Template.Spec.InitContainers

		failures = append(failures, resourcesFailuresForContainers("container", deployment.Namespace, deployment.Name, containers)...)
		failures = append(failures, resourcesFailuresForContainers("initContainer", deployment.Namespace, deployment.Name, initContainers)...)

		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", deployment.Namespace, deployment.Name)] = common.PreAnalysis{
				FailureDetails: failures,
				Deployment:     deployment,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, deployment.Name, deployment.Namespace).Set(float64(len(failures)))
		}
	}

	for key, value := range preAnalysis {
		a.Results = append(a.Results, common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		})
	}

	return a.Results, nil
}

func resourcesFailuresForContainers(containerType, namespace, deploymentName string, containers []corev1.Container) []common.Failure {
	var failures []common.Failure

	for _, c := range containers {
		missing := missingResourceFields(c)
		if len(missing) == 0 {
			continue
		}

		failures = append(failures, common.Failure{
			Text: fmt.Sprintf(
				"%s %s in Deployment %s/%s is missing resource settings: %s. Add CPU/memory requests and limits to improve scheduling stability and reduce eviction/noisy-neighbor issues.",
				containerType,
				c.Name,
				namespace,
				deploymentName,
				strings.Join(missing, ", "),
			),
			Sensitive: []common.Sensitive{
				{
					Unmasked: namespace,
					Masked:   util.MaskString(namespace),
				},
				{
					Unmasked: deploymentName,
					Masked:   util.MaskString(deploymentName),
				},
				{
					Unmasked: c.Name,
					Masked:   util.MaskString(c.Name),
				},
			},
		})
	}

	return failures
}

func missingResourceFields(c corev1.Container) []string {
	var missing []string

	req := c.Resources.Requests
	lim := c.Resources.Limits

	if req == nil || req.Cpu().IsZero() {
		missing = append(missing, "requests.cpu")
	}
	if req == nil || req.Memory().IsZero() {
		missing = append(missing, "requests.memory")
	}
	if lim == nil || lim.Cpu().IsZero() {
		missing = append(missing, "limits.cpu")
	}
	if lim == nil || lim.Memory().IsZero() {
		missing = append(missing, "limits.memory")
	}

	return missing
}
