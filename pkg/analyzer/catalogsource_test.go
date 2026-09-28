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

func TestCatalogSourceAnalyzer_UnhealthyState_ReturnsResult(t *testing.T) {
	cs := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "operators.coreos.com/v1alpha1",
			"kind":       "CatalogSource",
			"metadata": map[string]any{
				"name":      "broken-operators-external",
				"namespace": "openshift-marketplace",
			},
			"status": map[string]any{
				"connectionState": map[string]any{
					"lastObservedState": "TRANSIENT_FAILURE",
					"address":           "not-a-real-host.invalid:50051",
				},
			},
		},
	}

	listKinds := map[schema.GroupVersionResource]string{
		{Group: "operators.coreos.com", Version: "v1alpha1", Resource: "catalogsources"}: "CatalogSourceList",
	}
	scheme := runtime.NewScheme()
	dc := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, cs)

	a := common.Analyzer{
		Context: context.TODO(),
		Client:  &kubernetes.Client{DynamicClient: dc},
	}

	res, err := (CatalogSourceAnalyzer{}).Analyze(a)
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Kind != "CatalogSource" || !strings.Contains(res[0].Name, "openshift-marketplace/broken-operators-external") {
		t.Fatalf("unexpected result: %#v", res[0])
	}
	if len(res[0].Error) == 0 || !strings.Contains(res[0].Error[0].Text, "TRANSIENT_FAILURE") {
		t.Fatalf("expected TRANSIENT_FAILURE in message, got %#v", res[0].Error)
	}
}

func TestCatalogSourceAnalyzer_HealthyOrNoState_Ignored(t *testing.T) {
	// One READY (healthy), one with no status at all: both should be ignored.
	ready := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "operators.coreos.com/v1alpha1",
			"kind":       "CatalogSource",
			"metadata": map[string]any{
				"name":      "ready-operators",
				"namespace": "openshift-marketplace",
			},
			"status": map[string]any{
				"connectionState": map[string]any{
					"lastObservedState": "READY",
					"address":           "somewhere",
				},
			},
		},
	}
	nostate := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "operators.coreos.com/v1alpha1",
			"kind":       "CatalogSource",
			"metadata": map[string]any{
				"name":      "no-status-operators",
				"namespace": "openshift-marketplace",
			},
		},
	}

	listKinds := map[schema.GroupVersionResource]string{
		{Group: "operators.coreos.com", Version: "v1alpha1", Resource: "catalogsources"}: "CatalogSourceList",
	}
	scheme := runtime.NewScheme()
	dc := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, ready, nostate)

	a := common.Analyzer{
		Context: context.TODO(),
		Client:  &kubernetes.Client{DynamicClient: dc},
	}

	res, err := (CatalogSourceAnalyzer{}).Analyze(a)
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}
	if len(res) != 0 {
		t.Fatalf("expected 0 results (healthy/nostate ignored), got %d", len(res))
	}
}
