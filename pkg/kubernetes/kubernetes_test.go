/*
Copyright 2026 The K8sGPT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubernetes

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"

	policyreport "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/crd/api/policyreport/v1alpha2"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	gtwapi "sigs.k8s.io/gateway-api/apis/v1"
)

// writeTestKubeconfig stands up a stub API server that only answers the version
// discovery call NewClient makes, and writes a kubeconfig pointing at it.
func writeTestKubeconfig(t *testing.T) string {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(&version.Info{GitVersion: "v1.32.0"})
	}))
	t.Cleanup(server.Close)

	kubeconfig := filepath.Join(t.TempDir(), "config")
	require.NoError(t, clientcmd.WriteToFile(clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			"test": {Server: server.URL},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"test": {Cluster: "test"},
		},
		CurrentContext: "test",
	}, kubeconfig))

	return kubeconfig
}

// TestNewClientRegistersSchemes checks that the CRD types the analyzers need are
// on the shared scheme by the time NewClient returns. Registering them lazily
// from inside a concurrently-running Analyze() is what caused issue #1063.
//
// The built-in types are covered too: the client no longer uses
// controller-runtime's package-global default scheme, so nothing else puts
// them there.
func TestNewClientRegistersSchemes(t *testing.T) {
	client, err := NewClient("", writeTestKubeconfig(t))
	require.NoError(t, err)

	for _, tc := range []struct {
		obj runtime.Object
		gvk schema.GroupVersionKind
	}{
		{
			obj: &corev1.Pod{},
			gvk: corev1.SchemeGroupVersion.WithKind("Pod"),
		},
		{
			obj: &appsv1.Deployment{},
			gvk: appsv1.SchemeGroupVersion.WithKind("Deployment"),
		},
		{
			obj: &gtwapi.Gateway{},
			gvk: schema.GroupVersionKind{
				Group:   gtwapi.GroupVersion.Group,
				Version: gtwapi.GroupVersion.Version,
				Kind:    "Gateway",
			},
		},
		{
			obj: &policyreport.PolicyReport{},
			gvk: policyreport.SchemeGroupVersion.WithKind("PolicyReport"),
		},
		{
			obj: &policyreport.ClusterPolicyReport{},
			gvk: policyreport.SchemeGroupVersion.WithKind("ClusterPolicyReport"),
		},
	} {
		t.Run(tc.gvk.Kind, func(t *testing.T) {
			kinds, _, err := client.CtrlClient.Scheme().ObjectKinds(tc.obj)
			require.NoError(t, err)
			require.Contains(t, kinds, tc.gvk)
		})
	}
}

// TestNewClientDoesNotShareSchemeState is the other half of issue #1063.
// Moving registration out of Analyze() is not enough on its own: ctrl.New falls
// back to controller-runtime's package-global scheme when Options.Scheme is
// nil, so installing the CRD types from NewClient made client construction a
// writer of the very maps a running analysis reads through
// Scheme().ObjectKinds() on every List call. Building a fresh client while an
// existing one is in use crashed with "concurrent map read and map write".
func TestNewClientDoesNotShareSchemeState(t *testing.T) {
	kubeconfig := writeTestKubeconfig(t)

	// A client already built and mid-analysis while more are constructed.
	inUse, err := NewClient("", kubeconfig)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			if _, err := NewClient("", kubeconfig); err != nil {
				t.Errorf("NewClient returned an unexpected error: %v", err)
			}
		}()

		go func() {
			defer wg.Done()
			// The scheme lookup client.List performs before every request.
			if _, _, err := inUse.CtrlClient.Scheme().ObjectKinds(&policyreport.PolicyReport{}); err != nil {
				t.Errorf("ObjectKinds returned an unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	// Each client owns its scheme, so one cannot be written from under another.
	other, err := NewClient("", kubeconfig)
	require.NoError(t, err)
	require.NotSame(t, inUse.CtrlClient.Scheme(), other.CtrlClient.Scheme())
	require.NotSame(t, clientgoscheme.Scheme, inUse.CtrlClient.Scheme(),
		"client is still using the controller-runtime package-global scheme")
}

func TestNewClientReturnsSchemeInstallError(t *testing.T) {
	for name, swap := range map[string]func(func(*runtime.Scheme) error) func(){
		"client-go": func(stub func(*runtime.Scheme) error) func() {
			original := installClientGo
			installClientGo = stub
			return func() { installClientGo = original }
		},
		"gateway API": func(stub func(*runtime.Scheme) error) func() {
			original := installGatewayAPI
			installGatewayAPI = stub
			return func() { installGatewayAPI = original }
		},
		"policy report": func(stub func(*runtime.Scheme) error) func() {
			original := installPolicyReport
			installPolicyReport = stub
			return func() { installPolicyReport = original }
		},
	} {
		t.Run(name, func(t *testing.T) {
			kubeconfig := writeTestKubeconfig(t)

			sentinel := errors.New(name + " install failed")
			t.Cleanup(swap(func(*runtime.Scheme) error { return sentinel }))

			client, err := NewClient("", kubeconfig)
			require.Nil(t, client)
			require.ErrorIs(t, err, sentinel)
		})
	}
}
