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
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/magiconair/properties/assert"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"strings"
	"testing"
)

// sub-function
func badPod(name string, namespace string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
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
	}

}

// sub-function
func analysis_RunAnalysisFilterTester(t *testing.T, filterFlag string, specificResources []string) []common.Result {

	clientset := fake.NewSimpleClientset(
		badPod("example1", "default"),
		badPod("example2", "default"),
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
		Context:           context.Background(),
		Results:           []common.Result{},
		Namespace:         "default",
		MaxConcurrency:    1,
		SpecificResources: specificResources,
		Client: &kubernetes.Client{
			Client: clientset,
		},
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
	results = analysis_RunAnalysisFilterTester(t, "", []string{})
	assert.Equal(t, len(results), 4) // all built-in resource will be analyzed

	//2. When the --filter flag is specified

	filterFlag = "Pod" // --filter=Pod
	results = analysis_RunAnalysisFilterTester(t, filterFlag, []string{})
	assert.Equal(t, len(results), 2)
	assert.Equal(t, results[0].Kind, filterFlag)

	filterFlag = "Ingress,Pod" // --filter=Ingress,Pod
	results = analysis_RunAnalysisFilterTester(t, filterFlag, []string{})
	assert.Equal(t, len(results), 3)
}

// Test:  Filter logic with Active Filter
func TestAnalysis_RunAnalysisActiveFilter(t *testing.T) {

	//When the --filter flag is not specified but has actived filter in config
	var results []common.Result

	viper.SetDefault("active_filters", "Ingress")
	results = analysis_RunAnalysisFilterTester(t, "", []string{})
	assert.Equal(t, len(results), 1)

	viper.SetDefault("active_filters", []string{"Ingress", "Service"})
	results = analysis_RunAnalysisFilterTester(t, "", []string{})
	assert.Equal(t, len(results), 2)

	viper.SetDefault("active_filters", []string{"Ingress", "Service", "Pod"})
	results = analysis_RunAnalysisFilterTester(t, "", []string{})
	assert.Equal(t, len(results), 4)
}

// Test: test specific resources
func TestAnalysis_RunAnalysisWithSpecificRes_No_Filter(t *testing.T) {
	var results []common.Result

	viper.SetDefault("active_filters", []string{})
	// 1. one specific resource
	results = analysis_RunAnalysisFilterTester(t, "", []string{"pod/example1"})
	assert.Equal(t, len(results), 1)

	// 2. multiple specific resources
	results = analysis_RunAnalysisFilterTester(t, "", []string{"pod/example1", "service/example"})
	assert.Equal(t, len(results), 2)
	results = analysis_RunAnalysisFilterTester(t, "", []string{"pod/example1", "pod/example2", "service/example"})
	assert.Equal(t, len(results), 3)

	// 3. resource not found
	results = analysis_RunAnalysisFilterTester(t, "", []string{"pod/not-exist"})
	assert.Equal(t, len(results), 0)
}

// Test: test specific resources + filter
func TestAnalysis_RunAnalysisWithSpecificRes_With_Filter(t *testing.T) {
	var results []common.Result
	var filterFlag string
	// both `--filters` and specific resources are given

	// 4. single filter flag
	filterFlag = "Pod"
	//   4.1) non-kind resource
	results = analysis_RunAnalysisFilterTester(t, filterFlag, []string{"example1", "example2"})
	assert.Equal(t, len(results), 2)
	//   4.2) resource kind matches filter
	results = analysis_RunAnalysisFilterTester(t, filterFlag, []string{"pod/example1", "pod/example2"})
	assert.Equal(t, len(results), 2)
	//   4.3) resource kind not match filter, will be omitted
	results = analysis_RunAnalysisFilterTester(t, filterFlag, []string{"pod/example1", "service/example"})
	assert.Equal(t, len(results), 1)

	// 5. multiple filter flags
	filterFlag = "Pod,Service"
	//  5.1) kinds are all aligned
	results = analysis_RunAnalysisFilterTester(t, filterFlag, []string{"pod/example1", "pod/example2", "service/example"})
	assert.Equal(t, len(results), 3)
	//  5.2) some kinds are not aligned
	filterFlag = "Pod,Service,Deployment"
	results = analysis_RunAnalysisFilterTester(t, filterFlag, []string{"pod/example1", "pod/example2", "service/example"})
	assert.Equal(t, len(results), 3)

	// 6. invalid argument ( multiple filter flags but resources have no kind prefix)
	filterFlag = "Pod,Service"
	results = analysis_RunAnalysisFilterTester(t, filterFlag, []string{"example1"})
	assert.Equal(t, len(results), 0)
	results = analysis_RunAnalysisFilterTester(t, filterFlag, []string{"ping", "pong"})
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
