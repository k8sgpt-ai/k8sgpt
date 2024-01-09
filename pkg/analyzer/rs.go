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

type ReplicaSetAnalyzer struct{}

func (ReplicaSetAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "ReplicaSet"

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	// search all namespaces for pods that are not running
	list, err := a.Client.GetClient().AppsV1().ReplicaSets(a.Namespace).List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, rs := range list.Items {
		var failures []common.Failure

		// Check for empty rs
		if rs.Status.Replicas == 0 {

			// Check through container status to check for crashes
			for _, rsStatus := range rs.Status.Conditions {
				if rsStatus.Type == "ReplicaFailure" && rsStatus.Reason == "FailedCreate" {
					failures = append(failures, common.Failure{
						Text:      rsStatus.Message,
						Sensitive: []common.Sensitive{},
					})

				}
			}
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", rs.Namespace, rs.Name)] = common.PreAnalysis{
				ReplicaSet:     rs,
				FailureDetails: failures,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, rs.Name, rs.Namespace).Set(float64(len(failures)))
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.ReplicaSet.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}
	return a.Results, nil
}
