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
	"regexp"
	"strings"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	errorPattern = regexp.MustCompile(`(error|exception|fail)`)
	tailLines    = int64(100)
)

type LogAnalyzer struct {
}

func (LogAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "Log"

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	// search all namespaces for pods that are not running
	list, err := a.Client.GetClient().CoreV1().Pods(a.Namespace).List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var preAnalysis = map[string]common.PreAnalysis{}
	// Iterate through each pod

	for _, pod := range list.Items {
		podName := pod.Name
		for _, c := range pod.Spec.Containers {
			var failures []common.Failure
			podLogOptions := v1.PodLogOptions{
				TailLines: &tailLines,
				Container: c.Name,
			}
			podLogs, err := a.Client.Client.CoreV1().Pods(pod.Namespace).GetLogs(podName, &podLogOptions).DoRaw(a.Context)
			if err != nil {
				failures = append(failures, common.Failure{
					Text: fmt.Sprintf("Error %s from Pod %s", err.Error(), pod.Name),
					Sensitive: []common.Sensitive{
						{
							Unmasked: pod.Name,
							Masked:   util.MaskString(pod.Name),
						},
					},
				})
			} else {
				rawlogs := string(podLogs)
				if errorPattern.MatchString(strings.ToLower(rawlogs)) {
					failures = append(failures, common.Failure{
						Text: printErrorLines(pod.Name, pod.Namespace, rawlogs, errorPattern),
						Sensitive: []common.Sensitive{
							{
								Unmasked: pod.Name,
								Masked:   util.MaskString(pod.Name),
							},
						},
					})
				}
			}
			if len(failures) > 0 {
				preAnalysis[fmt.Sprintf("%s/%s/%s", pod.Namespace, pod.Name, c.Name)] = common.PreAnalysis{
					FailureDetails: failures,
					Pod:            pod,
				}
				AnalyzerErrorsMetric.WithLabelValues(kind, pod.Name, pod.Namespace).Set(float64(len(failures)))
			}
		}
	}
	for key, value := range preAnalysis {
		currentAnalysis := common.Result{
			Kind:  "Pod",
			Name:  key,
			Error: value.FailureDetails,
		}
		parent, _ := util.GetParent(a.Client, value.Pod.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}
func printErrorLines(podName, namespace, logs string, errorPattern *regexp.Regexp) string {
	// Split the logs into lines
	logLines := strings.Split(logs, "\n")

	// Check each line for errors and print the lines containing errors
	for _, line := range logLines {
		if errorPattern.MatchString(strings.ToLower(line)) {
			return line
		}
	}
	return ""
}
