package analysis

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/stretchr/testify/require"
)

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
