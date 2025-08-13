package analyzer

import (
	"context"
	"strings"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func TestSubscriptionAnalyzer(t *testing.T) {
	ok := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "operators.coreos.com/v1alpha1",
			"kind":       "Subscription",
			"metadata": map[string]any{
				"name":      "ok-sub",
				"namespace": "ns1",
			},
			"status": map[string]any{
				"state": "AtLatestKnown",
			},
		},
	}

	bad := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "operators.coreos.com/v1alpha1",
			"kind":       "Subscription",
			"metadata": map[string]any{
				"name":      "upgrade-sub",
				"namespace": "ns1",
			},
			"status": map[string]any{
				"state": "UpgradeAvailable",
				"conditions": []interface{}{
					map[string]any{
						"status":  "False",
						"reason":  "CatalogSourcesUnhealthy",
						"message": "not reachable",
					},
				},
			},
		},
	}

	listKinds := map[schema.GroupVersionResource]string{
		{Group: "operators.coreos.com", Version: "v1alpha1", Resource: "subscriptions"}: "SubscriptionList",
	}

	scheme := runtime.NewScheme()
	dc := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, ok, bad)

	a := common.Analyzer{
		Context: context.TODO(),
		Client:  &kubernetes.Client{DynamicClient: dc},
	}

	res, err := (SubscriptionAnalyzer{}).Analyze(a)
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Kind != "Subscription" || !strings.Contains(res[0].Name, "ns1/upgrade-sub") {
		t.Fatalf("unexpected result: %#v", res[0])
	}
	if len(res[0].Error) == 0 || !strings.Contains(res[0].Error[0].Text, "CatalogSourcesUnhealthy") {
		t.Fatalf("expected 'CatalogSourcesUnhealthy' in failure, got %#v", res[0].Error)
	}
}
