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

	list, err := a.Client.GetClient().CoreV1().Nodes().List(a.Context, metav1.ListOptions{})
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
				if nodeCondition.Status == v1.ConditionTrue {
					break
				}
				failures = addNodeConditionFailure(failures, node.Name, nodeCondition)
			default:
				if nodeCondition.Status != v1.ConditionFalse {
					failures = addNodeConditionFailure(failures, node.Name, nodeCondition)
				}
			}
		}

		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s", node.Name)] = common.PreAnalysis{
				Node:           node,
				FailureDetails: failures,
			}
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  "Node",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.Node.ObjectMeta)
		currentAnalysis.ParentObject = parent
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
				Masked:   util.MaskString(nodeName),
			},
		},
	})
	return failures
}
