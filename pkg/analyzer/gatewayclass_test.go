package analyzer

import (
	"context"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// A GatewayClass with no status conditions (newly created, or no controller
// installed) must not panic the analyzer.
func TestEmptyConditionsGatewayClassAnalyzer(t *testing.T) {
	GatewayClass := &gtwapi.GatewayClass{}
	GatewayClass.Name = "foobar"
	GatewayClass.Spec.ControllerName = "gateway.fooproxy.io/gatewayclass-controller"
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
	assert.Equal(t, len(analysisResults), 0)

}

func TestGatewayClassAnalyzerLabelSelectorFiltering(t *testing.T) {
	condition := metav1.Condition{
		Type:    "Accepted",
		Status:  "Ready",
		Message: "Ready",
		Reason:  "Ready",
	}

	// Create two GatewayClasses with different labels
	GatewayClass := &gtwapi.GatewayClass{}
	GatewayClass.Name = "foobar"
	GatewayClass.Spec.ControllerName = "gateway.fooproxy.io/gatewayclass-controller"
	GatewayClass.Labels = map[string]string{"app": "gatewayclass"}
	GatewayClass.Status.Conditions = []metav1.Condition{condition}

	GatewayClass2 := &gtwapi.GatewayClass{}
	GatewayClass2.Name = "foobar2"
	GatewayClass2.Spec.ControllerName = "gateway.fooproxy.io/gatewayclass-controller"
	GatewayClass2.Status.Conditions = []metav1.Condition{condition}

	scheme := scheme.Scheme
	err := gtwapi.Install(scheme)
	if err != nil {
		t.Error(err)
	}
	err = apiextensionsv1.AddToScheme(scheme)
	if err != nil {
		t.Error(err)
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(GatewayClass, GatewayClass2).Build()

	analyzerInstance := GatewayClassAnalyzer{}
	config := common.Analyzer{
		Client: &kubernetes.Client{
			CtrlClient: fakeClient,
		},
		Context:       context.Background(),
		Namespace:     "default",
		LabelSelector: "app=gatewayclass",
	}
	analysisResults, err := analyzerInstance.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 1)
}

func TestGatewayClassAnalyzerLabelSelectorOperators(t *testing.T) {
	condition := metav1.Condition{
		Type:    "Accepted",
		Status:  "Ready",
		Message: "Ready",
		Reason:  "Ready",
	}

	labeled := &gtwapi.GatewayClass{}
	labeled.Name = "labeled"
	labeled.Spec.ControllerName = "gateway.fooproxy.io/gatewayclass-controller"
	labeled.Labels = map[string]string{"app": "gatewayclass"}
	labeled.Status.Conditions = []metav1.Condition{condition}

	other := &gtwapi.GatewayClass{}
	other.Name = "other"
	other.Spec.ControllerName = "gateway.fooproxy.io/gatewayclass-controller"
	other.Labels = map[string]string{"app": "other"}
	other.Status.Conditions = []metav1.Condition{condition}

	scheme := scheme.Scheme
	require.NoError(t, gtwapi.Install(scheme))
	require.NoError(t, apiextensionsv1.AddToScheme(scheme))

	fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(labeled, other).Build()

	tests := []struct {
		selector string
		want     string
	}{
		{selector: "app==gatewayclass", want: "labeled"},
		{selector: "app!=gatewayclass", want: "other"},
		{selector: "app in (other)", want: "other"},
	}
	for _, tt := range tests {
		t.Run(tt.selector, func(t *testing.T) {
			config := common.Analyzer{
				Client: &kubernetes.Client{
					CtrlClient: fakeClient,
				},
				Context:       context.Background(),
				LabelSelector: tt.selector,
			}
			results, err := (GatewayClassAnalyzer{}).Analyze(config)
			require.NoError(t, err)
			require.Len(t, results, 1)
			assert.Equal(t, tt.want, results[0].Name)
		})
	}
}

func TestGatewayClassAnalyzer_MultipleConditionsOrdering(t *testing.T) {
	scheme := scheme.Scheme
	err := gtwapi.Install(scheme)
	if err != nil {
		t.Error(err)
	}
	err = apiextensionsv1.AddToScheme(scheme)
	if err != nil {
		t.Error(err)
	}

	// Case 1: SupportedVersion is at index 0 (Status: True), Accepted is at index 1 (Status: False).
	// Must report failure for Accepted condition even though index 0 is True.
	gcUnaccepted := &gtwapi.GatewayClass{}
	gcUnaccepted.Name = "unaccepted"
	gcUnaccepted.Spec.ControllerName = "gateway.fooproxy.io/gatewayclass-controller"
	gcUnaccepted.Status.Conditions = []metav1.Condition{
		{
			Type:    string(gtwapi.GatewayClassConditionStatusSupportedVersion),
			Status:  metav1.ConditionTrue,
			Message: "Supported version",
			Reason:  string(gtwapi.GatewayClassReasonSupportedVersion),
		},
		{
			Type:    string(gtwapi.GatewayClassConditionStatusAccepted),
			Status:  metav1.ConditionFalse,
			Message: "Controller rejected configuration",
			Reason:  string(gtwapi.GatewayClassReasonInvalidParameters),
		},
	}

	// Case 2: Another condition is at index 0 (Status: False), but Accepted is at index 1 (Status: True).
	// Must NOT report failure because Accepted is True.
	gcAccepted := &gtwapi.GatewayClass{}
	gcAccepted.Name = "accepted"
	gcAccepted.Spec.ControllerName = "gateway.fooproxy.io/gatewayclass-controller"
	gcAccepted.Status.Conditions = []metav1.Condition{
		{
			Type:    string(gtwapi.GatewayClassConditionStatusSupportedVersion),
			Status:  metav1.ConditionFalse,
			Message: "Unsupported version",
			Reason:  string(gtwapi.GatewayClassReasonUnsupportedVersion),
		},
		{
			Type:    string(gtwapi.GatewayClassConditionStatusAccepted),
			Status:  metav1.ConditionTrue,
			Message: "Controller accepted",
			Reason:  string(gtwapi.GatewayClassReasonAccepted),
		},
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(gcUnaccepted, gcAccepted).Build()

	analyzerInstance := GatewayClassAnalyzer{}
	config := common.Analyzer{
		Client: &kubernetes.Client{
			CtrlClient: fakeClient,
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	analysisResults, err := analyzerInstance.Analyze(config)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(analysisResults))
	assert.Equal(t, "unaccepted", analysisResults[0].Name)
	assert.Contains(t, analysisResults[0].Error[0].Text, "Controller rejected configuration")
}

