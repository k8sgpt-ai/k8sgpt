package analyzer

import (
	"fmt"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PdbAnalyzer struct{}

func (PdbAnalyzer) Analyze(a Analyzer) ([]Result, error) {

	list, err := a.Client.GetClient().PolicyV1().PodDisruptionBudgets(a.Namespace).List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]PreAnalysis{}

	for _, pdb := range list.Items {
		var failures []string

		evt, err := FetchLatestEvent(a.Context, a.Client, pdb.Namespace, pdb.Name)
		if err != nil || evt == nil {
			continue
		}

		if evt.Reason == "NoPods" && evt.Message != "" {
			if pdb.Spec.Selector != nil {
				for k, v := range pdb.Spec.Selector.MatchLabels {
					failures = append(failures, fmt.Sprintf("%s, expected label %s=%s", evt.Message, k, v))
				}
				for _, v := range pdb.Spec.Selector.MatchExpressions {
					failures = append(failures, fmt.Sprintf("%s, expected expression %s", evt.Message, v))
				}
			} else {
				failures = append(failures, fmt.Sprintf("%s, selector is nil", evt.Message))
			}
		}

		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", pdb.Namespace, pdb.Name)] = PreAnalysis{
				PodDisruptionBudget: pdb,
				FailureDetails:      failures,
			}
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = Result{
			Kind:  "PodDisruptionBudget",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.PodDisruptionBudget.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, err
}
