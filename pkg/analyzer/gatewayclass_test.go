package analyzer

import (
	"context"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gtwapi "sigs.k8s.io/gateway-api/apis/v1"
)

// Testing with the fake dynamic client if GatewayClasses have an accepted status
func TestGatewayClassAnalyzer(t *testing.T) {
	GatewayClass := &gtwapi.GatewayClass{}
	GatewayClass.Name = "foobar"
	GatewayClass.Spec.ControllerName = "gateway.fooproxy.io/gatewayclass-controller"
	// Initialize Conditions slice before setting properties
	BadCondition := metav1.Condition{
		Type:    "Accepted",
		Status:  "Uknown",
		Message: "Waiting for controller",
		Reason:  "Pending",
	}
	GatewayClass.Status.Conditions = []metav1.Condition{BadCondition}
	// Create a GatewayClassAnalyzer instance with the fake client
	scheme := scheme.Scheme
	err := gtwapi.Install(scheme)
	if err != nil {
		t.Error(err)
	}
	err = apiextensionsv1.AddToScheme(scheme)
	if err != nil {
		t.Error(err)
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(GatewayClass).Build()

	analyzerInstance := GatewayClassAnalyzer{}
	config := common.Analyzer{
		Client: &kubernetes.Client{
			CtrlClient: fakeClient,
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
