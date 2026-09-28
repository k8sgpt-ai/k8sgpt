package analyzer

import (
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		Resource(csvGVR).Namespace(metav1.NamespaceAll).
		List(a.Context, metav1.ListOptions{})
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
			// Superfície de condições para contexto
			if conds, _, _ := unstructured.NestedSlice(item.Object, "status", "conditions"); len(conds) > 0 {
				if msg := pickWorstCondition(conds); msg != "" {
					failures = append(failures, common.Failure{Text: fmt.Sprintf("phase=%q: %s", phase, msg)})
				}
			} else {
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

// reaproveitamos o heurístico já usado em outros pontos
func pickWorstCondition(conds []interface{}) string {
	for _, c := range conds {
		m, ok := c.(map[string]any)
		if !ok {
			continue
		}
		if s, _ := m["status"].(string); s == "True" {
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
