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

package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// newListEventsFixture starts an httptest server that stands in for the
// Kubernetes API. It applies the fieldSelector query parameter the same way
// the real API server would: as a filter over the full dataset, evaluated
// before Limit truncates the page. This lets the test tell apart a handler
// that sends involvedObject filters as a field selector from one that fetches
// an unfiltered page and filters it locally afterwards.
func newListEventsFixture(t *testing.T, namespace string, dataset []corev1.Event) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/version", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(&version.Info{GitVersion: "v1.32.0"})
	})
	mux.HandleFunc("/api/v1/namespaces/"+namespace+"/events", func(w http.ResponseWriter, r *http.Request) {
		items := dataset
		if fieldSelector := r.URL.Query().Get("fieldSelector"); fieldSelector != "" {
			sel, err := fields.ParseSelector(fieldSelector)
			require.NoError(t, err)
			items = nil
			for _, e := range dataset {
				if sel.Matches(fields.Set{
					"involvedObject.name": e.InvolvedObject.Name,
					"involvedObject.kind": e.InvolvedObject.Kind,
				}) {
					items = append(items, e)
				}
			}
		}
		if limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64); err == nil && limit > 0 && int64(len(items)) > limit {
			items = items[:limit]
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(&corev1.EventList{Items: items})
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func TestHandleListEventsAppliesInvolvedObjectFieldSelector(t *testing.T) {
	const namespace = "default"
	unrelated := corev1.Event{
		InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: "unrelated"},
	}
	matchingPod := corev1.Event{
		InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: "target"},
	}
	matchingDeployment := corev1.Event{
		InvolvedObject: corev1.ObjectReference{Kind: "Deployment", Name: "target"},
	}
	// The unrelated event is listed first so that, with Limit: 1, a handler
	// that filters the returned page locally (instead of sending the filter
	// to the API) would see only the unrelated event and return no matches.
	dataset := []corev1.Event{unrelated, matchingPod, matchingDeployment}

	fixture := newListEventsFixture(t, namespace, dataset)

	kubeconfig := filepath.Join(t.TempDir(), "config")
	require.NoError(t, clientcmd.WriteToFile(clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			"test": {Server: fixture.URL},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"test": {Cluster: "test"},
		},
		CurrentContext: "test",
	}, kubeconfig))
	t.Setenv("KUBECONFIG", kubeconfig)

	tests := []struct {
		name     string
		args     map[string]any
		expected []corev1.Event
	}{
		{
			name:     "name and kind filter returns the matching event beyond the first page",
			args:     map[string]any{"namespace": namespace, "involvedObjectName": "target", "involvedObjectKind": "Pod", "limit": int64(1)},
			expected: []corev1.Event{matchingPod},
		},
		{
			name:     "name only filter",
			args:     map[string]any{"namespace": namespace, "involvedObjectName": "target", "limit": int64(1)},
			expected: []corev1.Event{matchingPod},
		},
		{
			name:     "kind only filter",
			args:     map[string]any{"namespace": namespace, "involvedObjectKind": "Deployment", "limit": int64(1)},
			expected: []corev1.Event{matchingDeployment},
		},
		{
			name:     "no filter preserves existing unfiltered behavior",
			args:     map[string]any{"namespace": namespace, "limit": int64(1)},
			expected: []corev1.Event{unrelated},
		},
	}

	s := &K8sGptMCPServer{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := s.handleListEvents(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{Arguments: tt.args},
			})
			require.NoError(t, err)
			require.False(t, result.IsError, "unexpected tool error: %+v", result.Content)

			text, ok := mcp.AsTextContent(result.Content[0])
			require.True(t, ok, "expected text content")

			var got []corev1.Event
			require.NoError(t, json.Unmarshal([]byte(text.Text), &got))

			require.Len(t, got, len(tt.expected))
			for i, e := range tt.expected {
				require.Equal(t, e.InvolvedObject.Name, got[i].InvolvedObject.Name)
				require.Equal(t, e.InvolvedObject.Kind, got[i].InvolvedObject.Kind)
			}
		})
	}
}
