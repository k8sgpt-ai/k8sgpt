/*
Copyright 2025 The K8sGPT Authors.
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

package analyzer

import (
	"context"
	"sort"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestJobAnalyzer(t *testing.T) {
	tests := []struct {
		name         string
		config       common.Analyzer
		expectations []struct {
			name          string
			failuresCount int
		}
	}{
		{
			name: "Suspended Job",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&batchv1.Job{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "suspended-job",
								Namespace: "default",
							},
							Spec: batchv1.JobSpec{
								Suspend: boolPtr(true),
							},
						},
					),
				},
				Context:   context.Background(),
				Namespace: "default",
			},
			expectations: []struct {
				name          string
				failuresCount int
			}{
				{
					name:          "default/suspended-job",
					failuresCount: 1, // One failure for being suspended
				},
			},
		},

		{
			name: "Failed Job",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&batchv1.Job{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "failed-job",
								Namespace: "default",
							},
							Spec: batchv1.JobSpec{},
							Status: batchv1.JobStatus{
								Failed: 1,
							},
						},
					),
				},
				Context:   context.Background(),
				Namespace: "default",
			},
			expectations: []struct {
				name          string
				failuresCount int
			}{
				{
					name:          "default/failed-job",
					failuresCount: 1, // One failure for failed job
				},
			},
		},
		{
			name: "Valid Job",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&batchv1.Job{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "valid-job",
								Namespace: "default",
							},
							Spec: batchv1.JobSpec{},
						},
					),
				},
				Context:   context.Background(),
				Namespace: "default",
			},
			expectations: []struct {
				name          string
				failuresCount int
			}{
				// No expectations for valid job
			},
		},
		{
			name: "Multiple issues",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&batchv1.Job{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "multiple-issues",
								Namespace: "default",
							},
							Spec: batchv1.JobSpec{
								Suspend: boolPtr(true),
							},
							Status: batchv1.JobStatus{
								Failed: 1,
							},
						},
					),
				},
				Context:   context.Background(),
				Namespace: "default",
			},
			expectations: []struct {
				name          string
				failuresCount int
			}{
				{
					name:          "default/multiple-issues",
					failuresCount: 2, // Two failures: suspended and failed job
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := JobAnalyzer{}
			results, err := analyzer.Analyze(tt.config)
			require.NoError(t, err)
			require.Len(t, results, len(tt.expectations))

			// Sort results by name for consistent comparison
			sort.Slice(results, func(i, j int) bool {
				return results[i].Name < results[j].Name
			})

			for i, expectation := range tt.expectations {
				require.Equal(t, expectation.name, results[i].Name)
				require.Len(t, results[i].Error, expectation.failuresCount)
			}
		})
	}
}

func TestJobAnalyzerLabelSelector(t *testing.T) {
	clientSet := fake.NewSimpleClientset(
		&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "job-with-label",
				Namespace: "default",
				Labels: map[string]string{
					"app": "test",
				},
			},
			Spec: batchv1.JobSpec{},
			Status: batchv1.JobStatus{
				Failed: 1,
			},
		},
		&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "job-without-label",
				Namespace: "default",
			},
			Spec: batchv1.JobSpec{},
		},
	)

	// Test with label selector
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientSet,
		},
		Context:       context.Background(),
		Namespace:     "default",
		LabelSelector: "app=test",
	}

	analyzer := JobAnalyzer{}
	results, err := analyzer.Analyze(config)
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, "default/job-with-label", results[0].Name)
}
