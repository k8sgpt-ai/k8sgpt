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

	v1 "k8s.io/api/core/v1"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeAnalyzer struct{}

func (NodeAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "Node"

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	list, err := a.Client.GetClient().CoreV1().Nodes().List(a.Context, metav1.ListOptions{LabelSelector: a.LabelSelector})
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, node := range list.Items {
		var failures []common.Failure
		for _, nodeCondition := range node.Status.Conditions {
			// https://kubernetes.io/docs/concepts/architecture/nodes/#condition
			switch nodeCondition.Type {
			case v1.NodeReady:
				if nodeCondition.Status != v1.ConditionTrue {
					failures = addNodeConditionFailure(failures, node.Name, nodeCondition)
				}
			// k3s `EtcdIsVoter`` should not be reported as an error
			case v1.NodeConditionType("EtcdIsVoter"):
				break
			default:
				// For other conditions:
				// - Report True or Unknown status as failures (for standard conditions)
				// - Report any unknown condition type as a failure
				if nodeCondition.Status == v1.ConditionTrue || nodeCondition.Status == v1.ConditionUnknown || !isKnownNodeConditionType(nodeCondition.Type) {
					failures = addNodeConditionFailure(failures, node.Name, nodeCondition)
				}
			}
		}

		if len(failures) > 0 {
			preAnalysis[node.Name] = common.PreAnalysis{
				Node:           node,
				FailureDetails: failures,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, node.Name, "").Set(float64(len(failures)))

		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, found := util.GetParent(a.Client, value.Node.ObjectMeta)
		if found {
			currentAnalysis.ParentObject = parent
		}
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, err
}

func addNodeConditionFailure(failures []common.Failure, nodeName string, nodeCondition v1.NodeCondition) []common.Failure {
	failures = append(failures, common.Failure{
		Text: fmt.Sprintf("%s has condition of type %s, reason %s: %s", nodeName, nodeCondition.Type, nodeCondition.Reason, nodeCondition.Message),
		Sensitive: []common.Sensitive{
			{
				Unmasked: nodeName,
				Masked: func() string {
					masked, err := util.MaskString(nodeName)
					if err != nil {
						return nodeName
					}
					return masked
				}(),
			},
		},
	})
	return failures
}

// isKnownNodeConditionType checks if the condition type is a standard Kubernetes node condition
func isKnownNodeConditionType(conditionType v1.NodeConditionType) bool {
	switch conditionType {
	case v1.NodeReady,
		v1.NodeMemoryPressure,
		v1.NodeDiskPressure,
		v1.NodePIDPressure,
		v1.NodeNetworkUnavailable:
		return true
	default:
		return false
	}
}
