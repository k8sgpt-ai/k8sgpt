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

type PodAnalyzer struct {
}

func (PodAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "Pod"

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	// search all namespaces for pods that are not running
	list, err := a.Client.GetClient().CoreV1().Pods(a.Namespace).List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var preAnalysis = map[string]common.PreAnalysis{}

	for _, pod := range list.Items {
		var failures []common.Failure
		// Check for pending pods
		if pod.Status.Phase == "Pending" {

			// Check through container status to check for crashes
			for _, containerStatus := range pod.Status.Conditions {
				if containerStatus.Type == "PodScheduled" && containerStatus.Reason == "Unschedulable" {
					if containerStatus.Message != "" {
						failures = append(failures, common.Failure{
							Text:      containerStatus.Message,
							Sensitive: []common.Sensitive{},
						})
					}
				}
			}
		}

		// Check through container status to check for crashes or unready
		for _, containerStatus := range pod.Status.ContainerStatuses {

			if containerStatus.State.Waiting != nil {

				if isErrorReason(containerStatus.State.Waiting.Reason) && containerStatus.State.Waiting.Message != "" {
					failures = append(failures, common.Failure{
						Text:      containerStatus.State.Waiting.Message,
						Sensitive: []common.Sensitive{},
					})
				}

				// This represents a container that is still being created or blocked due to conditions such as OOMKilled
				if containerStatus.State.Waiting.Reason == "ContainerCreating" && pod.Status.Phase == "Pending" {

					// parse the event log and append details
					evt, err := FetchLatestEvent(a.Context, a.Client, pod.Namespace, pod.Name)
					if err != nil || evt == nil {
						continue
					}
					if isEvtErrorReason(evt.Reason) && evt.Message != "" {
						failures = append(failures, common.Failure{
							Text:      evt.Message,
							Sensitive: []common.Sensitive{},
						})
					}
				}

				// This represents container that is in CrashLoopBackOff state due to conditions such as OOMKilled
				if containerStatus.State.Waiting.Reason == "CrashLoopBackOff" {
					failures = append(failures, common.Failure{
						Text:      fmt.Sprintf("the last termination reason is %s container=%s pod=%s", containerStatus.LastTerminationState.Terminated.Reason, containerStatus.Name, pod.Name),
						Sensitive: []common.Sensitive{},
					})
				}
			} else {
				// when pod is Running but its ReadinessProbe fails
				if !containerStatus.Ready && pod.Status.Phase == "Running" {
					// parse the event log and append details
					evt, err := FetchLatestEvent(a.Context, a.Client, pod.Namespace, pod.Name)
					if err != nil || evt == nil {
						continue
					}
					if evt.Reason == "Unhealthy" && evt.Message != "" {
						failures = append(failures, common.Failure{
							Text:      evt.Message,
							Sensitive: []common.Sensitive{},
						})

					}

				}
			}
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)] = common.PreAnalysis{
				Pod:            pod,
				FailureDetails: failures,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, pod.Name, pod.Namespace).Set(float64(len(failures)))
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.Pod.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}

func isErrorReason(reason string) bool {
	failureReasons := []string{
		"CrashLoopBackOff", "ImagePullBackOff", "CreateContainerConfigError", "PreCreateHookError", "CreateContainerError",
		"PreStartHookError", "RunContainerError", "ImageInspectError", "ErrImagePull", "ErrImageNeverPull", "InvalidImageName",
	}

	for _, r := range failureReasons {
		if r == reason {
			return true
		}
	}
	return false
}

func isEvtErrorReason(reason string) bool {
	failureReasons := []string{
		"FailedCreatePodSandBox", "FailedMount",
	}

	for _, r := range failureReasons {
		if r == reason {
			return true
		}
	}
	return false
}
