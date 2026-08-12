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

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	gtwapi "sigs.k8s.io/gateway-api/apis/v1"
)

func TestNewClientRegistersGatewayAPIScheme(t *testing.T) {
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

	client, err := NewClient("", kubeconfig)
	require.NoError(t, err)

	kinds, _, err := client.CtrlClient.Scheme().ObjectKinds(&gtwapi.Gateway{})
	require.NoError(t, err)
	require.Contains(t, kinds, schema.GroupVersionKind{
		Group:   gtwapi.GroupVersion.Group,
		Version: gtwapi.GroupVersion.Version,
		Kind:    "Gateway",
	})
}

func TestNewClientReturnsGatewayAPIInstallError(t *testing.T) {
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

	sentinel := errors.New("gateway API install failed")
	original := installGatewayAPI
	installGatewayAPI = func(*runtime.Scheme) error { return sentinel }
	t.Cleanup(func() { installGatewayAPI = original })

	client, err := NewClient("", kubeconfig)
	require.Nil(t, client)
	require.ErrorIs(t, err, sentinel)
}
