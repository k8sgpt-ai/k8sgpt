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

func TestInstallPlanAnalyzer(t *testing.T) {
	ok := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "operators.coreos.com/v1alpha1",
			"kind":       "InstallPlan",
			"metadata": map[string]any{
				"name":      "ip-ok",
				"namespace": "ns1",
			},
			"status": map[string]any{"phase": "Complete"},
		},
	}

	bad := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "operators.coreos.com/v1alpha1",
			"kind":       "InstallPlan",
			"metadata": map[string]any{
				"name":      "ip-bad",
				"namespace": "ns1",
			},
			"status": map[string]any{
				"phase": "Failed",
				"conditions": []interface{}{
					map[string]any{
						"reason":  "ExecutionError",
						"message": "something went wrong",
					},
				},
			},
		},
	}

	listKinds := map[schema.GroupVersionResource]string{
		{Group: "operators.coreos.com", Version: "v1alpha1", Resource: "installplans"}: "InstallPlanList",
	}

	scheme := runtime.NewScheme()
	dc := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, ok, bad)

	a := common.Analyzer{
		Context: context.TODO(),
		Client:  &kubernetes.Client{DynamicClient: dc},
	}

	res, err := (InstallPlanAnalyzer{}).Analyze(a)
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Kind != "InstallPlan" || !strings.Contains(res[0].Name, "ns1/ip-bad") {
		t.Fatalf("unexpected result: %#v", res[0])
	}
	if len(res[0].Error) == 0 || !strings.Contains(res[0].Error[0].Text, "ExecutionError") {
		t.Fatalf("expected 'ExecutionError' in failure, got %#v", res[0].Error)
	}
}
