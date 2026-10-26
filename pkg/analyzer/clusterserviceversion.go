package analyzer

import (
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ClusterServiceVersionAnalyzer struct{}

var csvGVR = schema.GroupVersionResource{
	Group: "operators.coreos.com", Version: "v1alpha1", Resource: "clusterserviceversions",
}

func (ClusterServiceVersionAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	kind := "ClusterServiceVersion"

	if a.Client.GetDynamicClient() == nil {
		return nil, fmt.Errorf("dynamic client is nil in %s analyzer", kind)
	}

	list, err := a.Client.GetDynamicClient().
		Resource(csvGVR).Namespace(a.Namespace).
		List(a.Context, a.ListOptions())
	if err != nil {
		return nil, err
	}

	var results []common.Result
	for _, item := range list.Items {
		ns := item.GetNamespace()
		name := item.GetName()
		phase, _, _ := unstructured.NestedString(item.Object, "status", "phase")

		var failures []common.Failure
		if phase != "" && phase != "Succeeded" {
			// status.conditions is a history of phase changes, oldest first.
			// The current reason and message are kept on status itself.
			reason, _, _ := unstructured.NestedString(item.Object, "status", "reason")
			msg, _, _ := unstructured.NestedString(item.Object, "status", "message")
			switch {
			case reason != "" && msg != "":
				failures = append(failures, common.Failure{Text: fmt.Sprintf("phase=%q: %s: %s", phase, reason, msg)})
			case reason != "" || msg != "":
				failures = append(failures, common.Failure{Text: fmt.Sprintf("phase=%q: %s%s", phase, reason, msg)})
			default:
				failures = append(failures, common.Failure{Text: fmt.Sprintf("phase=%q (see status.conditions)", phase)})
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
