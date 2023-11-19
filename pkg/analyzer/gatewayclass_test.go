package analyzer

import (
	"context"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	clienttesting "k8s.io/client-go/testing"
)

func TestGatewayClassAnalyzer(t *testing.T) {
	unstructuredGatewayClass := map[string]interface{}{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "GatewayClass",
		"metadata": map[string]interface{}{
			"name": "foobar",
		},
		"spec": map[string]interface{}{
			"controllerName": "gateway.fooproxy.io/gatewayclass-controller",
		},
		"status": map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"message": "Valid GatewayClass",
					"reason":  "foo",
					"status":  "Uknown",
					"type":    "Accepted",
				},
			},
		},
	}
	mockGatewayClass := &unstructured.Unstructured{Object: unstructuredGatewayClass}

	// Create a mock unstructured list containing the mock GatewayClass object
	unstructuredList := &unstructured.UnstructuredList{}
	unstructuredList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "gateway.networking.k8s.io",
		Version: "v1",
		Kind:    "GatewayClassList",
	})
	unstructuredList.Items = []unstructured.Unstructured{*mockGatewayClass}

	fakeClient := fake.NewSimpleDynamicClient(runtime.NewScheme(), mockGatewayClass)
	// Inject mock data into the fake dynamic client
	fakeClient.PrependReactor("list", "gatewayclasses", func(action clienttesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, unstructuredList, nil
	})

	// Create a GatewayClassAnalyzer instance with the fake client
	analyzerInstance := GatewayClassAnalyzer{}
	config := common.Analyzer{
		Client: &kubernetes.Client{
			DynClient: fakeClient,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	analysisResults, err := analyzerInstance.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 1)

}
