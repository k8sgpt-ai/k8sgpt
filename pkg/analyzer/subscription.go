package analyzer

import (
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type SubscriptionAnalyzer struct{}

var subGVR = schema.GroupVersionResource{
	Group: "operators.coreos.com", Version: "v1alpha1", Resource: "subscriptions",
}

func (SubscriptionAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	kind := "Subscription"
	if a.Client.GetDynamicClient() == nil {
		return nil, fmt.Errorf("dynamic client is nil in %s analyzer", kind)
	}

	list, err := a.Client.GetDynamicClient().
		Resource(subGVR).Namespace(metav1.NamespaceAll).
		List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var results []common.Result
	for _, item := range list.Items {
		ns, name := item.GetNamespace(), item.GetName()
		state, _, _ := unstructured.NestedString(item.Object, "status", "state")
		conds, _, _ := unstructured.NestedSlice(item.Object, "status", "conditions")

		var failures []common.Failure
		if state == "" || state == "UpgradePending" || state == "UpgradeAvailable" {
			msg := "subscription not at latest"
			if c := pickWorstCondition(conds); c != "" {
				msg += "; " + c
			}
			failures = append(failures, common.Failure{Text: fmt.Sprintf("state=%q: %s", state, msg)})
		}

		if len(failures) > 0 {
			results = append(results, common.Result{
				Kind:  kind,
				Name:  ns + "/" + name,
				Error: failures,
			})
		}
	}
	return results, nil
}
