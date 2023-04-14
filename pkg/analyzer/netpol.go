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
		var failures []common.Failure

		// Check if policy allows traffic to all pods in the namespace
		if len(policy.Spec.PodSelector.MatchLabels) == 0 {
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf("Network policy allows traffic to all pods: %s", policy.Name),
				Sensitive: []common.Sensitive{
					{
						Unmasked: policy.Name,
						Masked:   util.MaskString(policy.Name),
					},
				},
			})
			continue
		}
		// Check if policy is not applied to any pods
		podList, err := util.GetPodListByLabels(a.Client.GetClient(), a.Namespace, policy.Spec.PodSelector.MatchLabels)
		if err != nil {
			return nil, err
		}
		if len(podList.Items) == 0 {
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf("Network policy is not applied to any pods: %s", policy.Name),
				Sensitive: []common.Sensitive{
					{
						Unmasked: policy.Name,
						Masked:   util.MaskString(policy.Name),
					},
				},
			})
		}

		if len(failures) > 0 {
			preAnalysis[policy.Name] = common.PreAnalysis{
				FailureDetails: failures,
				NetworkPolicy:  policy,
			}
		}
	}

	for key, value := range preAnalysis {
		currentAnalysis := common.Result{
			Kind:         "NetworkPolicy",
			Name:         key,
			Error:        value.FailureDetails,
			ParentObject: "",
		}
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}
