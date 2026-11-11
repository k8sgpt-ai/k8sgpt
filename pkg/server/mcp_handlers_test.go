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
	"slices"
	"strconv"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analysis"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
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

func TestFilterToolsStartFromCoreFilters(t *testing.T) {
	coreFilters, _, _ := analyzer.ListFilters()
	withoutPod := slices.DeleteFunc(slices.Clone(coreFilters), func(f string) bool { return f == "Pod" })

	s := &K8sGptMCPServer{}
	tests := []struct {
		name    string
		handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
		active  []string
		filters []string
		want    []string
	}{
		{
			name:    "add with no active filters keeps the core analyzers",
			handler: s.handleAddFilters,
			filters: []string{"HorizontalPodAutoscaler"},
			want:    append(slices.Clone(coreFilters), "HorizontalPodAutoscaler"),
		},
		{
			name:    "remove with no active filters removes from the core analyzers",
			handler: s.handleRemoveFilters,
			filters: []string{"Pod"},
			want:    withoutPod,
		},
		{
			name:    "add with active filters appends to them",
			handler: s.handleAddFilters,
			active:  []string{"Pod"},
			filters: []string{"HorizontalPodAutoscaler"},
			want:    []string{"Pod", "HorizontalPodAutoscaler"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)
			viper.SetConfigFile(filepath.Join(t.TempDir(), "config.yaml"))
			if tt.active != nil {
				viper.Set("active_filters", tt.active)
			}

			result, err := tt.handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{Arguments: map[string]any{"filters": tt.filters}},
			})
			require.NoError(t, err)
			require.False(t, result.IsError, "unexpected tool error: %+v", result.Content)
			require.ElementsMatch(t, tt.want, viper.GetStringSlice("active_filters"))
		})
	}
}

func TestHandleConfigCustomAnalyzersValidation(t *testing.T) {
	s := &K8sGptMCPServer{}

	t.Run("fails when custom analyzer name is empty", func(t *testing.T) {
		result, err := s.handleConfig(context.Background(), mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]any{
					"customAnalyzers": []map[string]any{
						{
							"name": "",
							"connection": map[string]any{
								"url":  "http://localhost",
								"port": 8080,
							},
						},
					},
				},
			},
		})
		require.NoError(t, err)
		require.True(t, result.IsError)
		text := result.Content[0].(mcp.TextContent).Text
		require.Contains(t, text, "custom analyzer name is required")
	})

	t.Run("fails when custom analyzer connection url is empty", func(t *testing.T) {
		result, err := s.handleConfig(context.Background(), mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]any{
					"customAnalyzers": []map[string]any{
						{
							"name": "test-analyzer",
							"connection": map[string]any{
								"url":  "",
								"port": 8080,
							},
						},
					},
				},
			},
		})
		require.NoError(t, err)
		require.True(t, result.IsError)
		text := result.Content[0].(mcp.TextContent).Text
		require.Contains(t, text, "custom analyzer connection url is required for test-analyzer")
	})

	t.Run("succeeds with valid custom analyzer", func(t *testing.T) {
		viper.Reset()
		t.Cleanup(viper.Reset)
		viper.SetConfigFile(filepath.Join(t.TempDir(), "config.yaml"))

		result, err := s.handleConfig(context.Background(), mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: map[string]any{
					"customAnalyzers": []map[string]any{
						{
							"name": "test-analyzer",
							"connection": map[string]any{
								"url":  "http://localhost",
								"port": 8080,
							},
						},
					},
				},
			},
		})
		require.NoError(t, err)
		require.False(t, result.IsError)
		text := result.Content[0].(mcp.TextContent).Text
		require.Contains(t, text, "Successfully added configuration")
	})
}

func TestHandleAddFilters_Validation(t *testing.T) {
	s := &K8sGptMCPServer{}

	tests := []struct {
		name        string
		args        map[string]any
		expectedErr string
	}{
		{
			name:        "empty filters array",
			args:        map[string]any{"filters": []string{}},
			expectedErr: "filters array is required and cannot be empty",
		},
		{
			name:        "empty filter name",
			args:        map[string]any{"filters": []string{""}},
			expectedErr: "filter name cannot be empty",
		},
		{
			name:        "non-existent filter",
			args:        map[string]any{"filters": []string{"NonExistentFilter"}},
			expectedErr: "Filter NonExistentFilter does not exist. Please use list-filters to view available filters",
		},
		{
			name:        "multiple non-existent filters",
			args:        map[string]any{"filters": []string{"FakeFilter1", "FakeFilter2"}},
			expectedErr: "Filter FakeFilter1, FakeFilter2 does not exist. Please use list-filters to view available filters",
		},
		{
			name:        "mixture of valid and non-existent filters",
			args:        map[string]any{"filters": []string{"Pod", "FakeFilter"}},
			expectedErr: "Filter FakeFilter does not exist. Please use list-filters to view available filters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)
			configPath := filepath.Join(t.TempDir(), "config.yaml")
			viper.SetConfigFile(configPath)
			viper.Set("active_filters", []string{"Service"})

			result, err := s.handleAddFilters(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{Arguments: tt.args},
			})
			require.NoError(t, err)
			require.True(t, result.IsError)
			text, ok := mcp.AsTextContent(result.Content[0])
			require.True(t, ok)
			require.Contains(t, text.Text, tt.expectedErr)
			// Ensure active_filters was not modified on error
			require.Equal(t, []string{"Service"}, viper.GetStringSlice("active_filters"))
		})
	}
}

func TestHandleAddFilters_AlreadyActive(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	viper.SetConfigFile(configPath)
	viper.Set("active_filters", []string{"Pod", "Service"})

	s := &K8sGptMCPServer{}
	result, err := s.handleAddFilters(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{"filters": []string{"Pod"}}},
	})
	require.NoError(t, err)
	require.False(t, result.IsError)
	text, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	require.Contains(t, text.Text, "All specified filters were already active")
}

func TestHandleRemoveFilters_Validation(t *testing.T) {
	s := &K8sGptMCPServer{}

	tests := []struct {
		name        string
		args        map[string]any
		expectedErr string
	}{
		{
			name:        "empty filters array",
			args:        map[string]any{"filters": []string{}},
			expectedErr: "filters array is required and cannot be empty",
		},
		{
			name:        "empty filter name",
			args:        map[string]any{"filters": []string{""}},
			expectedErr: "filter name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)
			configPath := filepath.Join(t.TempDir(), "config.yaml")
			viper.SetConfigFile(configPath)
			viper.Set("active_filters", []string{"Pod"})

			result, err := s.handleRemoveFilters(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{Arguments: tt.args},
			})
			require.NoError(t, err)
			require.True(t, result.IsError)
			text, ok := mcp.AsTextContent(result.Content[0])
			require.True(t, ok)
			require.Contains(t, text.Text, tt.expectedErr)
			require.Equal(t, []string{"Pod"}, viper.GetStringSlice("active_filters"))
		})
	}
}

func TestHandleRemoveFilters_NoneActive(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	viper.SetConfigFile(configPath)
	viper.Set("active_filters", []string{"Pod"})

	s := &K8sGptMCPServer{}
	result, err := s.handleRemoveFilters(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{"filters": []string{"Service"}}},
	})
	require.NoError(t, err)
	require.False(t, result.IsError)
	text, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	require.Contains(t, text.Text, "None of the specified filters were active")
}

// The MCP analyze tool is the last entry point that ran analysis without
// applying the exclusions config: a muted finding still reached MCP clients
// and, with Explain set, still cost an AI call. The CLI and the server-mode
// handler both apply exclusions; this asserts the MCP path does too.
func TestHandleAnalyzeAppliesExclusions(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "web", Namespace: "default"},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{{
				Name: "app",
				State: corev1.ContainerState{
					Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"},
				},
				LastTerminationState: corev1.ContainerState{
					Terminated: &corev1.ContainerStateTerminated{Reason: "OOMKilled", ExitCode: 137},
				},
			}},
		},
	}
	patches := gomonkey.ApplyFunc(kubernetes.NewClient, func(kubecontext, kubeconfig string) (*kubernetes.Client, error) {
		return &kubernetes.Client{
			Client: fake.NewSimpleClientset(&pod),
			Config: &rest.Config{Host: "fake-server"},
		}, nil
	})
	t.Cleanup(patches.Reset)

	viper.Set("verbose", false)
	viper.Set("custom_analyzers", []interface{}{})
	t.Cleanup(func() { viper.Set("exclusions", []interface{}{}) })

	s := &K8sGptMCPServer{aiProvider: &ai.AIProvider{}}
	request := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{"output": "json"}}}

	// Without exclusions the CrashLoopBackOff finding reaches the MCP client.
	result, err := s.handleAnalyze(context.Background(), request)
	require.NoError(t, err)
	require.False(t, result.IsError, "unexpected tool error: %+v", result.Content)
	text, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok, "expected text content")
	require.Contains(t, text.Text, "the last termination reason is OOMKilled")

	// Muting that finding's hash must remove it from the MCP output too.
	viper.Set("exclusions", []map[string]interface{}{
		{
			"reason": "OOMKilled on restarts is a known downstream effect of the JVM heap config",
			"hash":   analysis.FindingHash("Pod", "default/web", "the last termination reason is OOMKilled container=app pod=web"),
		},
	})
	result, err = s.handleAnalyze(context.Background(), request)
	require.NoError(t, err)
	require.False(t, result.IsError, "unexpected tool error: %+v", result.Content)
	text, ok = mcp.AsTextContent(result.Content[0])
	require.True(t, ok, "expected text content")
	require.NotContains(t, text.Text, "the last termination reason is OOMKilled")
}
