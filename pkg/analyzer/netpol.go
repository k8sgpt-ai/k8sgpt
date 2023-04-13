package analyzer

import (
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NetworkPolicyAnalyzer struct{}

func (NetworkPolicyAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	// get all network policies in the namespace
	policies, err := a.Client.GetClient().NetworkingV1().
		NetworkPolicies(a.Namespace).List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, policy := range policies.Items {
		// Check if policy allows traffic to all pods in the namespace
		if len(policy.Spec.PodSelector.MatchLabels) == 0 {
			preAnalysis[fmt.Sprintf("%s/%s", policy.Namespace, policy.Name)] = common.PreAnalysis{
				NetworkPolicy: policy,
				FailureDetails: []common.Failure{
					{
						Text: fmt.Sprintf("Network policy allows traffic to all pods in the namespace: %s", policy.Name),
					},
				},
			}
			continue
		}
		// Check if policy is not applied to any pods
		podList, err := util.GetPodListByLabels(a.Client.GetClient(), a.Namespace, policy.Spec.PodSelector.MatchLabels)
		if err != nil {
			return nil, err
		}
		if len(podList.Items) == 0 {
			preAnalysis[fmt.Sprintf("%s/%s", policy.Namespace, policy.Name)] = common.PreAnalysis{
				NetworkPolicy: policy,
				FailureDetails: []common.Failure{
					{
						Text: fmt.Sprintf("Network policy is not applied to any pods: %s", policy.Name),
					},
				},
			}
		}
	}

	var analysisResults []common.Result

	for key, value := range preAnalysis {
		currentAnalysis := common.Result{
			Kind:         "NetworkPolicy",
			Name:         key,
			Error:        value.FailureDetails,
			ParentObject: "",
		}
		analysisResults = append(analysisResults, currentAnalysis)
	}

	return analysisResults, nil
}
