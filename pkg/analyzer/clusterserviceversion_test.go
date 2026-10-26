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

func TestClusterServiceVersionAnalyzer(t *testing.T) {
	ok := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "operators.coreos.com/v1alpha1",
			"kind":       "ClusterServiceVersion",
			"metadata": map[string]any{
				"name":      "ok",
				"namespace": "ns1",
			},
			"status": map[string]any{"phase": "Succeeded"},
		},
	}

	bad := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "operators.coreos.com/v1alpha1",
			"kind":       "ClusterServiceVersion",
			"metadata": map[string]any{
				"name":      "bad",
				"namespace": "ns1",
			},
			"status": map[string]any{
				"phase":   "Failed",
				"reason":  "ErrorResolving",
				"message": "missing dep",
				// IMPORTANT: conditions must be []interface{}, not []map[string]any
				"conditions": []interface{}{
					map[string]any{
						"status":  "False",
						"reason":  "ErrorResolving",
						"message": "missing dep",
					},
				},
			},
		},
	}

	listKinds := map[schema.GroupVersionResource]string{
		{Group: "operators.coreos.com", Version: "v1alpha1", Resource: "clusterserviceversions"}: "ClusterServiceVersionList",
	}

	// Use a non-nil scheme with dynamicfake
	scheme := runtime.NewScheme()
	dc := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, ok, bad)

	a := common.Analyzer{
		Context: context.TODO(),
		Client:  &kubernetes.Client{DynamicClient: dc},
	}

	res, err := (ClusterServiceVersionAnalyzer{}).Analyze(a)
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Kind != "ClusterServiceVersion" || !strings.Contains(res[0].Name, "ns1/bad") {
		t.Fatalf("unexpected result: %#v", res[0])
	}
	if len(res[0].Error) == 0 || !strings.Contains(res[0].Error[0].Text, "missing dep") {
		t.Fatalf("expected 'missing dep' in failure, got %#v", res[0].Error)
	}
}

func TestClusterServiceVersionAnalyzerReportsCurrentReason(t *testing.T) {
	failed := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "operators.coreos.com/v1alpha1",
			"kind":       "ClusterServiceVersion",
			"metadata": map[string]any{
				"name":      "etcdoperator.v0.9.4",
				"namespace": "ns1",
			},
			"status": map[string]any{
				"phase":   "Failed",
				"reason":  "InstallCheckFailed",
				"message": "install timeout",
				"conditions": []interface{}{
					map[string]any{
						"phase":   "Pending",
						"reason":  "RequirementsUnknown",
						"message": "requirements not yet checked",
					},
					map[string]any{
						"phase":   "InstallReady",
						"reason":  "AllRequirementsMet",
						"message": "all requirements found, attempting install",
					},
					map[string]any{
						"phase":   "Installing",
						"reason":  "InstallSucceeded",
						"message": "waiting for install components to report healthy",
					},
					map[string]any{
						"phase":   "Failed",
						"reason":  "InstallCheckFailed",
						"message": "install timeout",
					},
				},
			},
		},
	}

	listKinds := map[schema.GroupVersionResource]string{
		{Group: "operators.coreos.com", Version: "v1alpha1", Resource: "clusterserviceversions"}: "ClusterServiceVersionList",
	}
	dc := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds, failed)

	a := common.Analyzer{
		Context: context.TODO(),
		Client:  &kubernetes.Client{DynamicClient: dc},
	}

	res, err := (ClusterServiceVersionAnalyzer{}).Analyze(a)
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}
	if len(res) != 1 || len(res[0].Error) != 1 {
		t.Fatalf("expected 1 result with 1 failure, got %#v", res)
	}
	want := `phase="Failed": InstallCheckFailed: install timeout`
	if got := res[0].Error[0].Text; got != want {
		t.Fatalf("failure text = %q, want %q", got, want)
	}
}
