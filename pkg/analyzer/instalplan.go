package analyzer

import (
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type InstallPlanAnalyzer struct{}

var ipGVR = schema.GroupVersionResource{
	Group: "operators.coreos.com", Version: "v1alpha1", Resource: "installplans",
}

func (InstallPlanAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	kind := "InstallPlan"
	if a.Client.GetDynamicClient() == nil {
		return nil, fmt.Errorf("dynamic client is nil in %s analyzer", kind)
	}

	list, err := a.Client.GetDynamicClient().
		Resource(ipGVR).Namespace(metav1.NamespaceAll).
		List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var results []common.Result
	for _, item := range list.Items {
		ns, name := item.GetNamespace(), item.GetName()
		phase, _, _ := unstructured.NestedString(item.Object, "status", "phase")

		var failures []common.Failure
		if phase != "" && phase != "Complete" {
			reason := firstCondStr(&item, "reason")
			msg := firstCondStr(&item, "message")
			switch {
			case reason != "" && msg != "":
				failures = append(failures, common.Failure{Text: fmt.Sprintf("phase=%q: %s: %s", phase, reason, msg)})
			case reason != "" || msg != "":
				failures = append(failures, common.Failure{Text: fmt.Sprintf("phase=%q: %s%s", phase, reason, msg)})
			default:
				failures = append(failures, common.Failure{Text: fmt.Sprintf("phase=%q (approval/manual? check status.conditions)", phase)})
			}
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

func firstCondStr(u *unstructured.Unstructured, field string) string {
	conds, _, _ := unstructured.NestedSlice(u.Object, "status", "conditions")
	if len(conds) == 0 {
		return ""
	}
	m, _ := conds[0].(map[string]any)
	if m == nil {
		return ""
	}
	v, _ := m[field].(string)
	return v
}
