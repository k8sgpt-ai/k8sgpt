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
	"testing"
	"time"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func stubFindings(t *testing.T, results []common.Result, errs []string) *[]string {
	t.Helper()
	var gotNamespaces []string
	orig := runFindingsAnalysis
	runFindingsAnalysis = func(namespace string, filters []string, labelSelector string) ([]common.Result, []string, error) {
		gotNamespaces = append(gotNamespaces, namespace)
		return results, errs, nil
	}
	t.Cleanup(func() { runFindingsAnalysis = orig })
	viper.Reset()
	t.Cleanup(viper.Reset)
	return &gotNamespaces
}

var sampleResults = []common.Result{
	{Kind: "Pod", Name: "default/web-1", ParentObject: "Deployment/web", Error: []common.Failure{
		{Text: "back-off restarting failed container", Sensitive: []common.Sensitive{{Unmasked: "web-1", Masked: "xxx"}}},
	}},
	{Kind: "Node", Name: "node-a", Error: []common.Failure{{Text: "node-a has condition of type Ready, reason KubeletNotReady"}}},
	{Kind: "Deployment", Name: "apps/api", Error: []common.Failure{{Text: "Deployment apps/api has 3 replicas but 1 are available"}}},
}

func TestBuildFindingsReport(t *testing.T) {
	now := time.Date(2026, 10, 30, 7, 0, 0, 0, time.UTC)
	got := buildFindingsReport("", []string{"Pod"}, sampleResults, []string{"Storage: forbidden"}, now)

	require.Equal(t, FindingsSchemaVersion, got.SchemaVersion)
	require.Equal(t, now, got.GeneratedAt)
	require.Equal(t, 3, got.Problems)
	require.Equal(t, []string{"Storage: forbidden"}, got.Errors)
	// Sorted by kind, then namespace, then name.
	require.Equal(t, []Finding{
		{Kind: "Deployment", Namespace: "apps", Name: "api", Errors: []string{"Deployment apps/api has 3 replicas but 1 are available"}},
		{Kind: "Node", Name: "node-a", Errors: []string{"node-a has condition of type Ready, reason KubeletNotReady"}},
		{Kind: "Pod", Namespace: "default", Name: "web-1", ParentObject: "Deployment/web", Errors: []string{"back-off restarting failed container"}},
	}, got.Findings)
}

func TestBuildFindingsReportEmpty(t *testing.T) {
	got := buildFindingsReport("default", nil, nil, nil, time.Now())
	data, err := json.Marshal(got)
	require.NoError(t, err)
	// An empty cluster still yields a list, never null, so clients can iterate.
	require.Contains(t, string(data), `"findings":[]`)
	require.Contains(t, string(data), `"problems":0`)
}

func TestHandleGetFindings(t *testing.T) {
	namespaces := stubFindings(t, sampleResults, nil)
	s := &K8sGptMCPServer{}

	result, err := s.handleGetFindings(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{"namespace": "default"}},
	})
	require.NoError(t, err)
	require.False(t, result.IsError, "unexpected tool error: %+v", result.Content)
	require.Equal(t, []string{"default"}, *namespaces)

	text, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	var report FindingsReport
	require.NoError(t, json.Unmarshal([]byte(text.Text), &report))
	require.Equal(t, FindingsSchemaVersion, report.SchemaVersion)
	require.Len(t, report.Findings, 3)
	require.NotContains(t, text.Text, "Unmasked", "findings must not leak sensitive fields")
}

func TestReadFindingsResource(t *testing.T) {
	namespaces := stubFindings(t, sampleResults, nil)
	s := &K8sGptMCPServer{}

	tests := []struct {
		uri       string
		namespace string
		wantErr   bool
	}{
		{uri: "k8sgpt://findings", namespace: ""},
		{uri: "k8sgpt://findings/kube-system", namespace: "kube-system"},
		{uri: "k8sgpt://findings/", wantErr: true},
		{uri: "k8sgpt://findings/a/b", wantErr: true},
		{uri: "k8sgpt://other", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			*namespaces = nil
			req := mcp.ReadResourceRequest{}
			req.Params.URI = tt.uri
			contents, err := s.readFindingsResource(context.Background(), req)
			if tt.wantErr {
				require.Error(t, err)
				require.Empty(t, *namespaces, "no analysis should run for an invalid URI")
				return
			}
			require.NoError(t, err)
			require.Equal(t, []string{tt.namespace}, *namespaces)
			require.Len(t, contents, 1)
			text, ok := contents[0].(*mcp.TextResourceContents)
			require.True(t, ok)
			require.Equal(t, tt.uri, text.URI)
			require.Equal(t, "application/json", text.MIMEType)
		})
	}
}

func TestRegisterFindings(t *testing.T) {
	s := &K8sGptMCPServer{server: server.NewMCPServer("test", "0.0.0",
		server.WithToolCapabilities(true), server.WithResourceCapabilities(true, false))}
	s.registerFindings()

	call := func(method string) string {
		resp := s.server.HandleMessage(context.Background(), []byte(`{"jsonrpc":"2.0","id":1,"method":"`+method+`"}`))
		data, err := json.Marshal(resp)
		require.NoError(t, err)
		return string(data)
	}
	tools := call("tools/list")
	require.Contains(t, tools, `"name":"get-findings"`)
	require.Contains(t, tools, `"readOnlyHint":true`)
	require.Contains(t, call("resources/list"), `"uri":"k8sgpt://findings"`)
	require.Contains(t, call("resources/templates/list"), `"uriTemplate":"k8sgpt://findings/{namespace}"`)
}
