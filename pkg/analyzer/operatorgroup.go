package analyzer

import (
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type OperatorGroupAnalyzer struct{}

var ogGVR = schema.GroupVersionResource{
	Group: "operators.coreos.com", Version: "v1", Resource: "operatorgroups",
}

func (OperatorGroupAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	kind := "OperatorGroup"
	if a.Client.GetDynamicClient() == nil {
		return nil, fmt.Errorf("dynamic client is nil in %s analyzer", kind)
	}

	list, err := a.Client.GetDynamicClient().
		Resource(ogGVR).Namespace(metav1.NamespaceAll).
		List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	countByNS := map[string]int{}
	for _, it := range list.Items {
		countByNS[it.GetNamespace()]++
	}

	var results []common.Result
	for ns, n := range countByNS {
		if n > 1 {
			results = append(results, common.Result{
				Kind:  kind,
				Name:  ns,
				Error: []common.Failure{{Text: fmt.Sprintf("%d OperatorGroups in namespace; this can break CSV resolution", n)}},
			})
		}
	}
	return results, nil
}
