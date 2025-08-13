package analyzer

import (
	"fmt"
	"strings"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type CatalogSourceAnalyzer struct{}

var catSrcGVR = schema.GroupVersionResource{
	Group:    "operators.coreos.com",
	Version:  "v1alpha1",
	Resource: "catalogsources",
}

func (CatalogSourceAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	kind := "CatalogSource"
	if a.Client.GetDynamicClient() == nil {
		return nil, fmt.Errorf("dynamic client is nil in %s analyzer", kind)
	}

	list, err := a.Client.GetDynamicClient().
		Resource(catSrcGVR).Namespace(metav1.NamespaceAll).
		List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var results []common.Result
	for _, item := range list.Items {
		ns, name := item.GetNamespace(), item.GetName()

		state, _, _ := unstructured.NestedString(item.Object, "status", "connectionState", "lastObservedState")
		addr, _, _ := unstructured.NestedString(item.Object, "status", "connectionState", "address")

		// Only report if state is present and not READY
		if state != "" && strings.ToUpper(state) != "READY" {
			results = append(results, common.Result{
				Kind: kind,
				Name: ns + "/" + name,
				Error: []common.Failure{{
					Text: fmt.Sprintf("connectionState=%s (address=%s)", state, addr),
				}},
			})
		}
	}
	return results, nil
}
