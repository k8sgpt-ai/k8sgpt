package rs

import (
	"fmt"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/common"

	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ReplicaSetAnalyzer struct {
	common.Analyzer
}

func (a *ReplicaSetAnalyzer) Analyze() error {
	// search all namespaces for pods that are not running
	list, err := a.Client.GetClient().AppsV1().ReplicaSets(a.Namespace).List(a.Context, metav1.ListOptions{})
	if err != nil {
		return err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, rs := range list.Items {
		var failures []string

		// Check for empty rs
		if rs.Status.Replicas == 0 {

			// Check through container status to check for crashes
			for _, rsStatus := range rs.Status.Conditions {
				if rsStatus.Type == "ReplicaFailure" && rsStatus.Reason == "FailedCreate" {
					failures = []string{rsStatus.Message}
				}
			}
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", rs.Namespace, rs.Name)] = common.PreAnalysis{
				ReplicaSet:     rs,
				FailureDetails: failures,
			}
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  "ReplicaSet",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.ReplicaSet.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Result = append(a.Result, currentAnalysis)
	}
	return nil
}

func (a *ReplicaSetAnalyzer) GetResult() []common.Result {
	return a.Result
}
