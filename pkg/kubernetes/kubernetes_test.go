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
	"testing"

	policyreport "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/crd/api/policyreport/v1alpha2"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
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
func TestNewClientRegistersSchemes(t *testing.T) {
	client, err := NewClient("", writeTestKubeconfig(t))
	require.NoError(t, err)

	for _, tc := range []struct {
		obj runtime.Object
		gvk schema.GroupVersionKind
	}{
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

func TestNewClientReturnsSchemeInstallError(t *testing.T) {
	for name, swap := range map[string]func(func(*runtime.Scheme) error) func(){
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
