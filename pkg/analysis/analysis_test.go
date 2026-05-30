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
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/cache"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	"github.com/magiconair/properties/assert"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

// helper function: get type name of an analyzer
func getTypeName(i interface{}) string {
	return reflect.TypeOf(i).Name()
}

// helper function: run analysis with filter
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

// Test: Verbose output in NewAnalysis with explain=false
func TestVerbose_NewAnalysisWithoutExplain(t *testing.T) {
	// Set viper config.
	viper.Set("verbose", true)
	viper.Set("kubecontext", "dummy")
	viper.Set("kubeconfig", "dummy")

	// Patch kubernetes.NewClient to return a dummy client.
	patches := gomonkey.ApplyFunc(kubernetes.NewClient, func(kubecontext, kubeconfig string) (*kubernetes.Client, error) {
		return &kubernetes.Client{
			Config: &rest.Config{Host: "fake-server"},
		}, nil
	})
	defer patches.Reset()

	output := util.CaptureOutput(func() {
		a, err := NewAnalysis(
			"", "english", []string{"Pod"}, "default", "", true,
			false, // explain
			10, false, false, []string{}, false,
		)
		require.NoError(t, err)
		a.Close()
	})

	expectedOutputs := []string{
		"Debug: Checking kubernetes client initialization.",
		"Debug: Kubernetes client initialized, server=fake-server.",
		"Debug: Checking cache configuration.",
		"Debug: Cache configuration loaded, type=file.",
		"Debug: Cache disabled.",
		"Debug: Analysis configuration loaded, filters=[Pod], language=english, namespace=default, labelSelector=none, explain=false, maxConcurrency=10, withDoc=false, withStats=false.",
	}
	for _, expected := range expectedOutputs {
		if !util.Contains(output, expected) {
			t.Errorf("Expected output to contain: '%s', but got output: '%s'", expected, output)
		}
	}
}

// Test: Verbose output in NewAnalysis with explain=true
func TestVerbose_NewAnalysisWithExplain(t *testing.T) {
	// Set viper config.
	viper.Set("verbose", true)
	viper.Set("kubecontext", "dummy")
	viper.Set("kubeconfig", "dummy")
	// Set a dummy AI configuration.
	dummyAIConfig := map[string]interface{}{
		"defaultProvider": "dummy",
		"providers": []map[string]interface{}{
			{
				"name":          "dummy",
				"baseUrl":       "http://dummy",
				"model":         "dummy-model",
				"customHeaders": map[string]string{},
			},
		},
	}
	viper.Set("ai", dummyAIConfig)

	// Patch kubernetes.NewClient to return a dummy client.
	patches := gomonkey.ApplyFunc(kubernetes.NewClient, func(kubecontext, kubeconfig string) (*kubernetes.Client, error) {
		return &kubernetes.Client{
			Config: &rest.Config{Host: "fake-server"},
		}, nil
	})
	defer patches.Reset()

	// Patch ai.NewClient to return a NoOp client.
	patches2 := gomonkey.ApplyFunc(ai.NewClient, func(name string) ai.IAI {
		return &ai.NoOpAIClient{}
	})
	defer patches2.Reset()

	output := util.CaptureOutput(func() {
		a, err := NewAnalysis(
			"", "english", []string{"Pod"}, "default", "", true,
			true, // explain
			10, false, false, []string{}, false,
		)
		require.NoError(t, err)
		a.Close()
	})

	expectedOutputs := []string{
		"Debug: Checking AI configuration.",
		"Debug: Using default AI provider dummy.",
		"Debug: AI configuration loaded, provider=dummy, baseUrl=http://dummy, model=dummy-model.",
		"Debug: Checking AI client initialization.",
		"Debug: AI client initialized.",
	}
	for _, expected := range expectedOutputs {
		if !util.Contains(output, expected) {
			t.Errorf("Expected output to contain: '%s', but got output: '%s'", expected, output)
		}
	}
}

// Test: Verbose output in RunAnalysis with filter flag
func TestVerbose_RunAnalysisWithFilter(t *testing.T) {
	viper.Set("verbose", true)
	// Run analysis with a filter flag ("Pod") to trigger debug output.
	output := util.CaptureOutput(func() {
		_ = analysis_RunAnalysisFilterTester(t, "Pod")
	})

	expectedOutputs := []string{
		"Debug: Filter flags [Pod] specified, run selected core analyzers.",
		"Debug: PodAnalyzer launched.",
		"Debug: PodAnalyzer completed without errors.",
	}

	for _, expected := range expectedOutputs {
		if !util.Contains(output, expected) {
			t.Errorf("Expected output to contain: '%s', but got output: '%s'", expected, output)
		}
	}
}

// Test: Verbose output in RunAnalysis with active filter
func TestVerbose_RunAnalysisWithActiveFilter(t *testing.T) {
	viper.Set("verbose", true)
	viper.SetDefault("active_filters", "Ingress")
	output := util.CaptureOutput(func() {
		_ = analysis_RunAnalysisFilterTester(t, "")
	})

	expectedOutputs := []string{
		"Debug: Found active filters [Ingress], run selected core analyzers.",
		"Debug: IngressAnalyzer launched.",
		"Debug: IngressAnalyzer completed without errors.",
	}

	for _, expected := range expectedOutputs {
		if !util.Contains(output, expected) {
			t.Errorf("Expected output to contain: '%s', but got output: '%s'", expected, output)
		}
	}
}

// Test: Verbose output in RunAnalysis without any filter (run all core analyzers)
func TestVerbose_RunAnalysisWithoutFilter(t *testing.T) {
	viper.Set("verbose", true)
	// Clear filter flag and active_filters to run all core analyzers.
	viper.SetDefault("active_filters", []string{})
	output := util.CaptureOutput(func() {
		_ = analysis_RunAnalysisFilterTester(t, "")
	})

	// Check for debug message indicating no filters.
	expectedNoFilter := "Debug: No filters selected and no active filters found, run all core analyzers."
	if !util.Contains(output, expectedNoFilter) {
		t.Errorf("Expected output to contain: '%s', but got output: '%s'", expectedNoFilter, output)
	}

	// Get all core analyzers from analyzer.GetAnalyzerMap()
	coreAnalyzerMap, _ := analyzer.GetAnalyzerMap()
	for _, analyzerInstance := range coreAnalyzerMap {
		analyzerType := getTypeName(analyzerInstance)
		expectedLaunched := fmt.Sprintf("Debug: %s launched.", analyzerType)
		expectedCompleted := fmt.Sprintf("Debug: %s completed without errors.", analyzerType)
		if !util.Contains(output, expectedLaunched) {
			t.Errorf("Expected output to contain: '%s', but got output: '%s'", expectedLaunched, output)
		}
		if !util.Contains(output, expectedCompleted) {
			t.Errorf("Expected output to contain: '%s', but got output: '%s'", expectedCompleted, output)
		}
	}
}

// Test: Verbose output in RunCustomAnalysis without custom analyzer
func TestVerbose_RunCustomAnalysisWithoutCustomAnalyzer(t *testing.T) {
	viper.Set("verbose", true)
	// Set custom_analyzers to empty array to trigger "No custom analyzers" debug message.
	viper.Set("custom_analyzers", []interface{}{})
	analysisObj := &Analysis{
		MaxConcurrency: 1,
	}
	output := util.CaptureOutput(func() {
		analysisObj.RunCustomAnalysis()
	})
	expected := "Debug: No custom analyzers found."
	if !util.Contains(output, "Debug: No custom analyzers found.") {
		t.Errorf("Expected output to contain: '%s', but got output: '%s'", expected, output)
	}
}

// Test: Verbose output in RunCustomAnalysis with custom analyzer
func TestVerbose_RunCustomAnalysisWithCustomAnalyzer(t *testing.T) {
	viper.Set("verbose", true)
	// Set custom_analyzers with one custom analyzer using "fake" connection.
	viper.Set("custom_analyzers", []map[string]interface{}{
		{
			"name":       "TestCustomAnalyzer",
			"connection": map[string]interface{}{"url": "127.0.0.1", "port": "2333"},
		},
	})

	analysisObj := &Analysis{
		MaxConcurrency: 1,
	}
	output := util.CaptureOutput(func() {
		analysisObj.RunCustomAnalysis()
	})
	assert.Equal(t, 1, len(analysisObj.Errors)) // connection error

	expectedOutputs := []string{
		"Debug: Found custom analyzers [TestCustomAnalyzer].",
	}

	unexpectedOutputs := []string{
		"Debug: TestCustomAnalyzer launched.",
		"Debug: TestCustomAnalyzer completed with errors.",
	}


	for _, expected := range expectedOutputs {
		if !util.Contains(output, expected) {
			t.Errorf("Expected output to contain: '%s', but got output: '%s'", expected, output)
		}
	}

	for _, unexpected := range unexpectedOutputs {
		if util.Contains(output, unexpected) {
			t.Errorf("Did not expect output to contain: '%s', but it did. Full output: '%s'", unexpected, output)
		}
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
	output := util.CaptureOutput(func() {
		_ = analysisObj.GetAIResults("json", false)
	})

	expected := "Debug: Generating AI analysis."
	if !util.Contains(output, expected) {
		t.Errorf("Expected output to contain: '%s', but got output: '%s'", expected, output)
	}
}
