package analyzer

import (
	"context"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/magiconair/properties/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gtwapi "sigs.k8s.io/gateway-api/apis/v1"
)

func BuildGatewayClass(name string) gtwapi.GatewayClass {
	GatewayClass := gtwapi.GatewayClass{}
	GatewayClass.Name = name
	// Namespace is not needed outside of this test, GatewayClass is cluster-scoped
	GatewayClass.Namespace = "default"
	GatewayClass.Spec.ControllerName = "gateway.fooproxy.io/gatewayclass-controller"

	return GatewayClass
}

func BuildGateway(className gtwapi.ObjectName, status metav1.ConditionStatus) gtwapi.Gateway {
	Gateway := gtwapi.Gateway{}
	Gateway.Name = "foobar"
	Gateway.Namespace = "default"
	Gateway.Spec.GatewayClassName = className
	Gateway.Spec.Listeners = []gtwapi.Listener{
		{
			Name:     "proxy",
			Port:     80,
			Protocol: gtwapi.HTTPProtocolType,
		},
	}
	Condition := metav1.Condition{
		Type:    "Accepted",
		Status:  status,
		Message: "An expected message",
		Reason:  "Test",
	}
	Gateway.Status.Conditions = []metav1.Condition{Condition}

	return Gateway
}

func TestGatewayAnalyzer(t *testing.T) {
	ClassName := gtwapi.ObjectName("exists")
	AcceptedStatus := metav1.ConditionTrue
	GatewayClass := BuildGatewayClass(string(ClassName))

	Gateway := BuildGateway(ClassName, AcceptedStatus)
	// Create a Gateway Analyzer instance with the fake client
	scheme := scheme.Scheme

	err := gtwapi.Install(scheme)
	if err != nil {
		t.Error(err)
	}
	err = apiextensionsv1.AddToScheme(scheme)
	if err != nil {
		t.Error(err)
	}
	objects := []runtime.Object{
		&Gateway,
		&GatewayClass,
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()

	analyzerInstance := GatewayAnalyzer{}
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
	assert.Equal(t, len(analysisResults), 0)

}

func TestMissingClassGatewayAnalyzer(t *testing.T) {
	ClassName := gtwapi.ObjectName("non-existed")
	AcceptedStatus := metav1.ConditionTrue
	Gateway := BuildGateway(ClassName, AcceptedStatus)

	// Create a Gateway Analyzer instance with the fake client
	scheme := scheme.Scheme
	err := gtwapi.Install(scheme)
	if err != nil {
		t.Error(err)
	}
	err = apiextensionsv1.AddToScheme(scheme)
	if err != nil {
		t.Error(err)
	}
	objects := []runtime.Object{
		&Gateway,
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()

	analyzerInstance := GatewayAnalyzer{}
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

func TestStatusGatewayAnalyzer(t *testing.T) {
	ClassName := gtwapi.ObjectName("exists")
	AcceptedStatus := metav1.ConditionUnknown
	GatewayClass := BuildGatewayClass(string(ClassName))

	Gateway := BuildGateway(ClassName, AcceptedStatus)

	// Create a Gateway Analyzer instance with the fake client
	scheme := scheme.Scheme
	err := gtwapi.Install(scheme)
	if err != nil {
		t.Error(err)
	}
	err = apiextensionsv1.AddToScheme(scheme)
	if err != nil {
		t.Error(err)
	}
	objects := []runtime.Object{
		&Gateway,
		&GatewayClass,
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()

	analyzerInstance := GatewayAnalyzer{}
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
	var errorFound bool
	want := "Gateway 'default/foobar' is not accepted. Message: 'An expected message'."
	for _, analysis := range analysisResults {
		for _, got := range analysis.Error {
			if want == got.Text {
				errorFound = true
			}
		}
		if errorFound {
			break
		}
	}

	if !errorFound {
		t.Errorf("Expected message, <%v> , not found in Gateway's analysis results", want)
	}
}
