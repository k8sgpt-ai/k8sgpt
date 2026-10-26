package analyzer

import (
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
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
		Resource(subGVR).Namespace(a.Namespace).
		List(a.Context, a.ListOptions())
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

// pickWorstCondition returns the reason and message of the first True
// condition. Subscription conditions are abnormal-true, so a True condition
// such as ResolutionFailed is the problem and a False one is healthy.
func pickWorstCondition(conds []interface{}) string {
	for _, c := range conds {
		m, ok := c.(map[string]any)
		if !ok {
			continue
		}
		if s, _ := m["status"].(string); s != "True" {
			continue
		}
		r, _ := m["reason"].(string)
		msg, _ := m["message"].(string)
		if r == "" && msg == "" {
			continue
		}
		if r != "" && msg != "" {
			return r + ": " + msg
		}
		return r + msg
	}
	return ""
}
