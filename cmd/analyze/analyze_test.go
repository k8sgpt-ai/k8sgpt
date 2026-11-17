/*
Copyright 2023 The K8sGPT Authors.
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

package analyze

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

// --resource hands the name to every analyzer that runs, so the filter set and
// the resource kind must agree; a mismatched --filter would make unrelated
// analyzers select metadata.name for a kind the user never asked about.
func TestResolveResourceSelection(t *testing.T) {
	tests := []struct {
		name         string
		resource     string
		filters      []string
		wantFilters  []string
		wantResource string
		wantErr      string
	}{
		{
			name:        "no resource flag leaves filters untouched",
			filters:     []string{"Deployment", "Service"},
			wantFilters: []string{"Deployment", "Service"},
		},
		{
			name:         "resource alone selects its kind",
			resource:     "Deployment/my-app",
			wantFilters:  []string{"Deployment"},
			wantResource: "my-app",
		},
		{
			name:         "matching filter reduces to the resource kind",
			resource:     "Deployment/my-app",
			filters:      []string{"Deployment"},
			wantFilters:  []string{"Deployment"},
			wantResource: "my-app",
		},
		{
			name:         "matching filter is case-insensitive",
			resource:     "Deployment/my-app",
			filters:      []string{"deployment"},
			wantFilters:  []string{"Deployment"},
			wantResource: "my-app",
		},
		{
			name:     "filter for a different kind is rejected",
			resource: "Deployment/my-app",
			filters:  []string{"Service"},
			wantErr:  `also requests "Service"`,
		},
		{
			name:     "multi-kind filter set is rejected, not silently narrowed",
			resource: "Deployment/my-app",
			filters:  []string{"Deployment", "Service"},
			wantErr:  `also requests "Service"`,
		},
		{
			name:     "missing name half is rejected",
			resource: "Deployment",
			wantErr:  "Kind/Name format",
		},
		{
			name:     "missing kind half is rejected",
			resource: "/my-app",
			wantErr:  "Kind/Name format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFilters, gotResource, err := resolveResourceSelection(tt.resource, tt.filters)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotResource != tt.wantResource {
				t.Errorf("resource name = %q, want %q", gotResource, tt.wantResource)
			}
			if len(gotFilters) != len(tt.wantFilters) {
				t.Fatalf("filters = %v, want %v", gotFilters, tt.wantFilters)
			}
			for i := range gotFilters {
				if gotFilters[i] != tt.wantFilters[i] {
					t.Errorf("filters = %v, want %v", gotFilters, tt.wantFilters)
				}
			}
		})
	}
}

// With --verbose the Debug lines go to stderr, so `analyze -o json --verbose`
// still writes a single JSON document to stdout.
func TestAnalyzeVerboseJSONKeepsStdoutClean(t *testing.T) {
	viper.Set("verbose", true)
	t.Cleanup(func() { viper.Set("verbose", false) })
	viper.Set("ai", map[string]interface{}{
		"providers": []map[string]interface{}{
			{"name": "openai", "baseUrl": "http://dummy", "model": "dummy-model", "customHeaders": map[string]string{}},
		},
	})
	t.Cleanup(func() { viper.Set("ai", nil) })
	oldFilters, oldOutput, oldNocache, oldConcurrency := filters, output, nocache, maxConcurrency
	oldExplain, oldCustom, oldSelector := explain, customAnalysis, labelSelector
	t.Cleanup(func() {
		filters, output, nocache, maxConcurrency = oldFilters, oldOutput, oldNocache, oldConcurrency
		explain, customAnalysis, labelSelector = oldExplain, oldCustom, oldSelector
	})
	filters, output, nocache, maxConcurrency = []string{"Pod"}, "json", true, 1
	labelSelector = "app=web"
	explain, customAnalysis = true, true

	patches := gomonkey.ApplyFunc(kubernetes.NewClient, func(kubecontext, kubeconfig string) (*kubernetes.Client, error) {
		return &kubernetes.Client{Client: fake.NewSimpleClientset(), Config: &rest.Config{Host: "fake-server"}}, nil
	})
	defer patches.Reset()
	patches2 := gomonkey.ApplyFunc(ai.NewClient, func(name string) ai.IAI { return &ai.NoOpAIClient{} })
	defer patches2.Reset()

	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w
	defer func() { os.Stderr = oldStderr }()

	stdout := util.CaptureOutput(func() {
		AnalyzeCmd.Run(AnalyzeCmd, nil)
	})
	require.NoError(t, w.Close())
	var stderr bytes.Buffer
	_, err = stderr.ReadFrom(r)
	require.NoError(t, err)

	require.True(t, json.Valid([]byte(stdout)), "stdout must hold only the JSON document, got %q", stdout)
	require.Contains(t, stderr.String(), "Debug: Checking output.")
}
