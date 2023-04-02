package analyzer

import (
	"context"
	"fmt"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AnalyzePdb(ctx context.Context, config *AnalysisConfiguration, client *kubernetes.Client, aiClient ai.IAI,
	analysisResults *[]Analysis) error {

	list, err := client.GetClient().PolicyV1().PodDisruptionBudgets(config.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	var preAnalysis = map[string]PreAnalysis{}

	for _, pdb := range list.Items {
		var failures []string

		evt, err := FetchLatestEvent(ctx, client, pdb.Namespace, pdb.Name)
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
		var currentAnalysis = Analysis{
			Kind:  "PodDisruptionBudget",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(client, value.PodDisruptionBudget.ObjectMeta)
		currentAnalysis.ParentObject = parent
		*analysisResults = append(*analysisResults, currentAnalysis)
	}

	return nil
}
