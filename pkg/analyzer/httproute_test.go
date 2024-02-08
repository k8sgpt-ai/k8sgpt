package analyzer

import (
	"context"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gtwapi "sigs.k8s.io/gateway-api/apis/v1"
)

func BuildRouteGateway(namespace, name, fromNamespaceref string) gtwapi.Gateway {
	routeNamespace := &gtwapi.RouteNamespaces{}
	switch fromNamespaceref {
	case "Same":
		fromSame := gtwapi.NamespacesFromSame
		routeNamespace.From = &fromSame
	case "Selector":
		fromSelector := gtwapi.NamespacesFromSelector
		routeNamespace.From = &fromSelector
		routeNamespace.Selector = &metav1.LabelSelector{}
		routeNamespace.Selector.MatchLabels = map[string]string{"foo": "bar"}

	default:
		fromAll := gtwapi.NamespacesFromAll
		routeNamespace.From = &fromAll
	}
	Gateway := gtwapi.Gateway{}
	Gateway.Name = name
	Gateway.Namespace = namespace
	Gateway.Spec.GatewayClassName = "fooclassName"
	Gateway.Spec.Listeners = []gtwapi.Listener{
		{
			Name:     "proxy",
			Port:     80,
			Protocol: gtwapi.HTTPProtocolType,
			AllowedRoutes: &gtwapi.AllowedRoutes{
				Namespaces: routeNamespace,
			},
		},
	}
	Condition := metav1.Condition{
		Type:    "Accepted",
		Status:  "True",
		Message: "An expected message",
		Reason:  "Test",
	}
	Gateway.Status.Conditions = []metav1.Condition{Condition}

	return Gateway
}

func BuildHTTPRoute(backendName, gtwName gtwapi.ObjectName, gtwNamespace gtwapi.Namespace, svcPort *gtwapi.PortNumber, namespace string) gtwapi.HTTPRoute {
	HTTPRoute := gtwapi.HTTPRoute{}
	HTTPRoute.Name = "foohttproute"
	HTTPRoute.Namespace = namespace
	HTTPRoute.Spec.ParentRefs = []gtwapi.ParentReference{
		{
			Name:      gtwName,
			Namespace: &gtwNamespace,
		},
	}
	HTTPRoute.Spec.Rules = []gtwapi.HTTPRouteRule{
		{
			BackendRefs: []gtwapi.HTTPBackendRef{
				{
					BackendRef: gtwapi.BackendRef{
						BackendObjectReference: gtwapi.BackendObjectReference{
							Name: backendName,
							Port: svcPort,
						},
					},
				},
			},
		},
	}
	return HTTPRoute
}

/*
	Testing different cases

1. Gateway doesn't exist or at least doesn't exist in the same namespace
2. Gateway exists in different namespace, is configured in httproute's spec
and Gateway's configuration is allowing only from its same namespace
3. Gateway exists in the same namespace but has selectors different from route's labels
4. BackendRef is pointing to a non existent Service
5. BackendRef's port and Service Port are different
*/
func TestGWMissiningHTTRouteAnalyzer(t *testing.T) {
	backendName := gtwapi.ObjectName("foobackend")
	gtwName := gtwapi.ObjectName("non-existent")
	gtwNamespace := gtwapi.Namespace("non-existent")
	svcPort := gtwapi.PortNumber(1027)
	httpRouteNamespace := "default"

	HTTPRoute := BuildHTTPRoute(backendName, gtwName, gtwNamespace, &svcPort, httpRouteNamespace)
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
		&HTTPRoute,
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()

	analyzerInstance := HTTPRouteAnalyzer{}
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
	want := "HTTPRoute uses the Gateway 'non-existent/non-existent' which does not exist in the same namespace."
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
		t.Errorf("Expected message, <%s> , not found in HTTPRoute's analysis results", want)
	}

}

func TestGWConfigSameHTTRouteAnalyzer(t *testing.T) {
	backendName := gtwapi.ObjectName("foobackend")
	gtwName := gtwapi.ObjectName("gatewayname")
	gtwNamespace := gtwapi.Namespace("differentnamespace")
	svcPort := gtwapi.PortNumber(1027)
	httpRouteNamespace := "default"

	HTTPRoute := BuildHTTPRoute(backendName, gtwName, gtwNamespace, &svcPort, httpRouteNamespace)

	Gateway := BuildRouteGateway("differentnamespace", "gatewayname", "Same")
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
		&HTTPRoute,
		&Gateway,
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()

	analyzerInstance := HTTPRouteAnalyzer{}
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
	want := "HTTPRoute 'default/foohttproute' is deployed in a different namespace from Gateway 'differentnamespace/gatewayname' which only allows HTTPRoutes from its namespace."
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
		t.Errorf("Expected message, <%s> , not found in HTTPRoute's analysis results", want)
	}
}
func TestGWConfigSelectorHTTRouteAnalyzer(t *testing.T) {
	backendName := gtwapi.ObjectName("foobackend")
	gtwName := gtwapi.ObjectName("gatewayname")
	gtwNamespace := gtwapi.Namespace("default")
	svcPort := gtwapi.PortNumber(1027)
	httpRouteNamespace := "default"

	HTTPRoute := BuildHTTPRoute(backendName, gtwName, gtwNamespace, &svcPort, httpRouteNamespace)

	Gateway := BuildRouteGateway("default", "gatewayname", "Selector")
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
		&HTTPRoute,
		&Gateway,
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()

	analyzerInstance := HTTPRouteAnalyzer{}
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
	want := "HTTPRoute 'default/foohttproute' can't be attached on Gateway 'default/gatewayname', selector labels do not match HTTProute's labels."
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
		t.Errorf("Expected message, <%s> , not found in HTTPRoute's analysis results", want)
	}
}

func TestSvcMissingHTTRouteAnalyzer(t *testing.T) {
	backendName := gtwapi.ObjectName("foobackend")
	gtwName := gtwapi.ObjectName("gatewayname")
	gtwNamespace := gtwapi.Namespace("default")
	svcPort := gtwapi.PortNumber(1027)
	httpRouteNamespace := "default"

	HTTPRoute := BuildHTTPRoute(backendName, gtwName, gtwNamespace, &svcPort, httpRouteNamespace)

	Gateway := BuildRouteGateway("default", "gatewayname", "Same")
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
		&HTTPRoute,
		&Gateway,
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()

	analyzerInstance := HTTPRouteAnalyzer{}
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
	want := "HTTPRoute uses the Service 'default/foobackend' which does not exist."
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
		t.Errorf("Expected message, <%s> , not found in HTTPRoute's analysis results", want)
	}
}
func TestSvcDifferentPortHTTRouteAnalyzer(t *testing.T) {
	//Add a Service Object
	Service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foobackend",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "example-app",
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
	backendName := gtwapi.ObjectName("foobackend")
	gtwName := gtwapi.ObjectName("gatewayname")
	gtwNamespace := gtwapi.Namespace("default")
	// different port
	svcPort := gtwapi.PortNumber(1027)
	httpRouteNamespace := "default"

	HTTPRoute := BuildHTTPRoute(backendName, gtwName, gtwNamespace, &svcPort, httpRouteNamespace)

	Gateway := BuildRouteGateway("default", "gatewayname", "Same")
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
		&HTTPRoute,
		&Gateway,
		&Service,
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()

	analyzerInstance := HTTPRouteAnalyzer{}
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
	want := "HTTPRoute's backend service 'foobackend' is using port '1027' but the corresponding K8s service 'default/foobackend' isn't configured with the same port."
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
		t.Errorf("Expected message, <%s> , not found in HTTPRoute's analysis results", want)
	}
}
