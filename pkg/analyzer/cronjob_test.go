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

func TestCronJobAnalyzer(t *testing.T) {
	tests := []struct {
		name         string
		config       common.Analyzer
		expectations []struct {
			name          string
			failuresCount int
		}
	}{
		{
			name: "Suspended CronJob",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&batchv1.CronJob{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "suspended-job",
								Namespace: "default",
							},
							Spec: batchv1.CronJobSpec{
								Schedule: "*/5 * * * *",
								Suspend:  boolPtr(true),
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
			name: "Invalid schedule format",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&batchv1.CronJob{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "invalid-schedule",
								Namespace: "default",
							},
							Spec: batchv1.CronJobSpec{
								Schedule: "invalid-cron", // Invalid cron format
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
					name:          "default/invalid-schedule",
					failuresCount: 1, // One failure for invalid schedule
				},
			},
		},
		{
			name: "Negative starting deadline",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&batchv1.CronJob{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "negative-deadline",
								Namespace: "default",
							},
							Spec: batchv1.CronJobSpec{
								Schedule:                "*/5 * * * *",
								StartingDeadlineSeconds: int64Ptr(-60), // Negative deadline
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
					name:          "default/negative-deadline",
					failuresCount: 1, // One failure for negative deadline
				},
			},
		},
		{
			name: "Valid CronJob",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&batchv1.CronJob{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "valid-job",
								Namespace: "default",
							},
							Spec: batchv1.CronJobSpec{
								Schedule: "*/5 * * * *", // Valid cron format
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
				// No expectations for valid job
			},
		},
		{
			name: "Multiple issues",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&batchv1.CronJob{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "multiple-issues",
								Namespace: "default",
							},
							Spec: batchv1.CronJobSpec{
								Schedule:                "invalid-cron",
								StartingDeadlineSeconds: int64Ptr(-60),
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
					failuresCount: 2, // Two failures: invalid schedule and negative deadline
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := CronJobAnalyzer{}
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

func TestCronJobAnalyzerLabelSelector(t *testing.T) {
	clientSet := fake.NewSimpleClientset(
		&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "job-with-label",
				Namespace: "default",
				Labels: map[string]string{
					"app": "test",
				},
			},
			Spec: batchv1.CronJobSpec{
				Schedule: "invalid-cron", // This should trigger a failure
			},
		},
		&batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "job-without-label",
				Namespace: "default",
			},
			Spec: batchv1.CronJobSpec{
				Schedule: "invalid-cron", // This should trigger a failure
			},
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

	analyzer := CronJobAnalyzer{}
	results, err := analyzer.Analyze(config)
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, "default/job-with-label", results[0].Name)
}

func TestCheckCronScheduleIsValid(t *testing.T) {
	tests := []struct {
		name     string
		schedule string
		wantErr  bool
	}{
		{
			name:     "Valid schedule - every 5 minutes",
			schedule: "*/5 * * * *",
			wantErr:  false,
		},
		{
			name:     "Valid schedule - specific time",
			schedule: "0 2 * * *",
			wantErr:  false,
		},
		{
			name:     "Valid schedule - complex",
			schedule: "0 0 1,15 * 3",
			wantErr:  false,
		},
		{
			name:     "Invalid schedule - wrong format",
			schedule: "invalid-cron",
			wantErr:  true,
		},
		{
			name:     "Invalid schedule - too many fields",
			schedule: "* * * * * *",
			wantErr:  true,
		},
		{
			name:     "Invalid schedule - empty string",
			schedule: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CheckCronScheduleIsValid(tt.schedule)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func int64Ptr(i int64) *int64 {
	return &i
}
