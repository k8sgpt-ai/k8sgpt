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

package analysis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/cache"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/magiconair/properties/assert"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// helper function to capture stdout
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = old
	return buf.String()
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

// sub-function
func analysis_RunAnalysisFilterTester(t *testing.T, filterFlag string) []common.Result {
	clientset := fake.NewSimpleClientset(
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example",
				Namespace: "default",
			},
			Status: v1.PodStatus{
				Phase: v1.PodPending,
				Conditions: []v1.PodCondition{
					{
						Type:    v1.PodScheduled,
						Reason:  "Unschedulable",
						Message: "0/1 nodes are available: 1 node(s) had taint {node-role.kubernetes.io/master: }, that the pod didn't tolerate.",
					},
				},
			},
		},
		&v1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		},
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: v1.ServiceSpec{
				Selector: map[string]string{
					"app": "example",
				},
			},
		},
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		},
	)

	analysis := Analysis{
		Context:        context.Background(),
		Results:        []common.Result{},
		Namespace:      "default",
		MaxConcurrency: 1,
		Client: &kubernetes.Client{
			Client: clientset,
		},
		WithDoc: true,
	}
	if len(filterFlag) > 0 {
		// `--filter` is explicitly given
		analysis.Filters = strings.Split(filterFlag, ",")
	}
	analysis.RunAnalysis()
	return analysis.Results

}

// Test: Filter logic with running different Analyzers
func TestAnalysis_RunAnalysisWithFilter(t *testing.T) {
	var results []common.Result
	var filterFlag string

	//1. Neither --filter flag Nor active filter is specified, only the "core analyzers"
	results = analysis_RunAnalysisFilterTester(t, "")
	assert.Equal(t, len(results), 3) // all built-in resource will be analyzed

	//2. When the --filter flag is specified

	filterFlag = "Pod" // --filter=Pod
	results = analysis_RunAnalysisFilterTester(t, filterFlag)
	assert.Equal(t, len(results), 1)
	assert.Equal(t, results[0].Kind, filterFlag)

	filterFlag = "Ingress,Pod" // --filter=Ingress,Pod
	results = analysis_RunAnalysisFilterTester(t, filterFlag)
	assert.Equal(t, len(results), 2)
}

// Test:  Filter logic with Active Filter
func TestAnalysis_RunAnalysisActiveFilter(t *testing.T) {

	//When the --filter flag is not specified but has actived filter in config
	var results []common.Result

	viper.SetDefault("active_filters", "Ingress")
	results = analysis_RunAnalysisFilterTester(t, "")
	assert.Equal(t, len(results), 1)

	viper.SetDefault("active_filters", []string{"Ingress", "Service"})
	results = analysis_RunAnalysisFilterTester(t, "")
	assert.Equal(t, len(results), 2)

	viper.SetDefault("active_filters", []string{"Ingress", "Service", "Pod"})
	results = analysis_RunAnalysisFilterTester(t, "")
	assert.Equal(t, len(results), 3)

	// Invalid filter
	results = analysis_RunAnalysisFilterTester(t, "invalid")
	assert.Equal(t, len(results), 0)
}

func TestAnalysis_NoProblemJsonOutput(t *testing.T) {

	analysis := Analysis{
		Results:   []common.Result{},
		Namespace: "default",
	}

	expected := JsonOutput{
		Status:   StateOK,
		Problems: 0,
		Results:  []common.Result{},
	}

	gotJson, err := analysis.PrintOutput("json")
	if err != nil {
		t.Error(err)
	}

	got := JsonOutput{}
	err = json.Unmarshal(gotJson, &got)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(got)
	fmt.Println(expected)

	require.Equal(t, got, expected)

}

func TestAnalysis_ProblemJsonOutput(t *testing.T) {
	analysis := Analysis{
		Results: []common.Result{
			{
				Kind: "Deployment",
				Name: "test-deployment",
				Error: []common.Failure{
					{
						Text:      "test-problem",
						Sensitive: []common.Sensitive{},
					},
				},
				Details:      "test-solution",
				ParentObject: "parent-resource"},
		},
		Namespace: "default",
	}

	expected := JsonOutput{
		Status:   StateProblemDetected,
		Problems: 1,
		Results: []common.Result{
			{
				Kind: "Deployment",
				Name: "test-deployment",
				Error: []common.Failure{
					{
						Text:      "test-problem",
						Sensitive: []common.Sensitive{},
					},
				},
				Details:      "test-solution",
				ParentObject: "parent-resource"},
		},
	}

	gotJson, err := analysis.PrintOutput("json")
	if err != nil {
		t.Error(err)
	}

	got := JsonOutput{}
	err = json.Unmarshal(gotJson, &got)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(got)
	fmt.Println(expected)

	require.Equal(t, got, expected)
}

func TestAnalysis_MultipleProblemJsonOutput(t *testing.T) {
	analysis := Analysis{
		Results: []common.Result{
			{
				Kind: "Deployment",
				Name: "test-deployment",
				Error: []common.Failure{
					{
						Text:      "test-problem",
						Sensitive: []common.Sensitive{},
					},
					{
						Text:      "another-test-problem",
						Sensitive: []common.Sensitive{},
					},
				},
				Details:      "test-solution",
				ParentObject: "parent-resource"},
		},
		Namespace: "default",
	}

	expected := JsonOutput{
		Status:   StateProblemDetected,
		Problems: 2,
		Results: []common.Result{
			{
				Kind: "Deployment",
				Name: "test-deployment",
				Error: []common.Failure{
					{
						Text:      "test-problem",
						Sensitive: []common.Sensitive{},
					},
					{
						Text:      "another-test-problem",
						Sensitive: []common.Sensitive{},
					},
				},
				Details:      "test-solution",
				ParentObject: "parent-resource"},
		},
	}

	gotJson, err := analysis.PrintOutput("json")
	if err != nil {
		t.Error(err)
	}

	got := JsonOutput{}
	err = json.Unmarshal(gotJson, &got)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(got)
	fmt.Println(expected)

	require.Equal(t, got, expected)
}

func TestNewAnalysis(t *testing.T) {
	disabledCache := cache.New("disabled-cache")
	disabledCache.DisableCache()
	aiClient := &ai.NoOpAIClient{}
	results := []common.Result{
		{
			Kind: "VulnerabilityReport",
			Error: []common.Failure{
				{
					Text:          "This is a custom failure",
					KubernetesDoc: "test-kubernetes-doc",
					Sensitive: []common.Sensitive{
						{
							Masked:   "masked-error",
							Unmasked: "unmasked-error",
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name        string
		a           Analysis
		output      string
		anonymize   bool
		expectedErr string
	}{
		{
			name: "Empty results",
			a:    Analysis{},
		},
		{
			name: "cache disabled",
			a: Analysis{
				AIClient: aiClient,
				Cache:    disabledCache,
				Results:  results,
			},
		},
		{
			name: "output and anonymize both set",
			a: Analysis{
				AIClient: aiClient,
				Cache:    cache.New("test-cache"),
				Results:  results,
			},
			output:    "test-output",
			anonymize: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.a.GetAIResults(tt.output, tt.anonymize)
			if tt.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.expectedErr)
			}
		})
	}
}

func TestGetAIResultForSanitizedFailures(t *testing.T) {
	enabledCache := cache.New("enabled-cache")
	disabledCache := cache.New("disabled-cache")
	disabledCache.DisableCache()
	aiClient := &ai.NoOpAIClient{}

	tests := []struct {
		name           string
		a              Analysis
		texts          []string
		promptTmpl     string
		expectedOutput string
		expectedErr    string
	}{
		{
			name: "Cache enabled",
			a: Analysis{
				AIClient: aiClient,
				Cache:    enabledCache,
			},
			texts:          []string{"some-data"},
			expectedOutput: "I am a noop response to the prompt %!(EXTRA string=, string=some-data)",
		},
		{
			name: "cache disabled",
			a: Analysis{
				AIClient: aiClient,
				Cache:    disabledCache,
				Language: "English",
			},
			texts:          []string{"test input"},
			promptTmpl:     "Response in %s: %s",
			expectedOutput: "I am a noop response to the prompt Response in English: test input",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			output, err := tt.a.getAIResultForSanitizedFailures(tt.texts, tt.promptTmpl)
			if tt.expectedErr == "" {
				require.NoError(t, err)
				require.Equal(t, tt.expectedOutput, output)
			} else {
				require.ErrorContains(t, err, tt.expectedErr)
				require.Empty(t, output)
			}
		})
	}
}

// Test: Verbose output in RunAnalysis with filter flag
func TestVerbose_RunAnalysisWithFilter(t *testing.T) {
	viper.Set("verbose", true)
	// Run analysis with a filter flag ("Pod") to trigger debug output.
	output := captureOutput(func() {
		_ = analysis_RunAnalysisFilterTester(t, "Pod")
	})
	if !contains(output, "Debug: Filter flags [Pod] specified, run selected core analyzers.") {
		t.Errorf("Expected debug output indicating filter flags [Pod] specified, but got: %s", output)
	}
	if !contains(output, "Debug: PodAnalyzer launched.") {
		t.Errorf("Expected debug output indicating PodAnalyzer launch, but got: %s", output)
	}
	if !contains(output, "Debug: PodAnalyzer completed without errors.") {
		t.Errorf("Expected debug output indicating PodAnalyzer completion without errors, but got: %s", output)
	}
}

// Test: Verbose output in RunAnalysis with active filter
func TestVerbose_RunAnalysisActiveFilter(t *testing.T) {
	viper.Set("verbose", true)
	viper.SetDefault("active_filters", "Ingress")
	output := captureOutput(func() {
		_ = analysis_RunAnalysisFilterTester(t, "")
	})
	if !contains(output, "Debug: Found active filters [Ingress], run selected core analyzers.") {
		t.Errorf("Expected debug output indicating active filters [Ingress] found, but got: %s", output)
	}
	if !contains(output, "Debug: IngressAnalyzer launched.") {
		t.Errorf("Expected debug output indicating IngressAnalyzer launch, but got: %s", output)
	}
	if !contains(output, "Debug: IngressAnalyzer completed without errors.") {
		t.Errorf("Expected debug output indicating IngressAnalyzer completion without errors, but got: %s", output)
	}
}

// Test: Verbose output in GetAIResults
func TestVerbose_GetAIResults(t *testing.T) {
	viper.Set("verbose", true)
	disabledCache := cache.New("disabled-cache")
	disabledCache.DisableCache()
	aiClient := &ai.NoOpAIClient{}
	analysisObj := Analysis{
		AIClient: aiClient,
		Cache:    disabledCache,
		Results: []common.Result{
			{
				Kind:         "Deployment",
				Name:         "test-deployment",
				Error:        []common.Failure{{Text: "test-problem", Sensitive: []common.Sensitive{}}},
				Details:      "test-solution",
				ParentObject: "parent-resource",
			},
		},
		Namespace: "default",
	}
	output := captureOutput(func() {
		_ = analysisObj.GetAIResults("json", false)
	})
	if !contains(output, "Debug: Generating AI analysis.") {
		t.Errorf("Expected debug output indicating AI analysis generation, but got: %s", output)
	}
}

// Test: Verbose output in RunCustomAnalysis
func TestVerbose_RunCustomAnalysis(t *testing.T) {
	viper.Set("verbose", true)
	// Set custom_analyzers to empty array to trigger "No custom analyzers" debug message.
	viper.Set("custom_analyzers", []interface{}{})
	analysisObj := &Analysis{
		MaxConcurrency: 1,
	}
	output := captureOutput(func() {
		analysisObj.RunCustomAnalysis()
	})
	if !contains(output, "Debug: No custom analyzers found.") {
		t.Errorf("Expected debug output indicating no custom analyzers found, but got: %s", output)
	}
}
