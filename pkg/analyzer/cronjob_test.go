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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCronJobAnalyzer(t *testing.T) {
	suspend := new(bool)
	*suspend = true

	invalidStartingDeadline := new(int64)
	*invalidStartingDeadline = -7

	validStartingDeadline := new(int64)
	*validStartingDeadline = 7

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: fake.NewSimpleClientset(
				&batchv1.CronJob{
					ObjectMeta: metav1.ObjectMeta{
						Name: "CJ1",
						// This CronJob won't be list because of namespace filtering.
						Namespace: "test",
					},
				},
				&batchv1.CronJob{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "CJ2",
						Namespace: "default",
					},
					// A suspended CronJob will contribute to failures.
					Spec: batchv1.CronJobSpec{
						Suspend: suspend,
					},
				},
				&batchv1.CronJob{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "CJ3",
						Namespace: "default",
					},
					Spec: batchv1.CronJobSpec{
						// Valid schedule
						Schedule: "*/1 * * * *",

						// Negative starting deadline
						StartingDeadlineSeconds: invalidStartingDeadline,
					},
				},
				&batchv1.CronJob{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "CJ4",
						Namespace: "default",
					},
					Spec: batchv1.CronJobSpec{
						// Invalid schedule
						Schedule: "*** * * * *",
					},
				},
				&batchv1.CronJob{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "CJ5",
						Namespace: "default",
					},
					Spec: batchv1.CronJobSpec{
						// Valid schedule
						Schedule: "*/1 * * * *",

						// Positive starting deadline shouldn't be any problem.
						StartingDeadlineSeconds: validStartingDeadline,
					},
				},
				&batchv1.CronJob{
					// This cronjob shouldn't contribute to any failures.
					ObjectMeta: metav1.ObjectMeta{
						Name:      "successful-cronjob",
						Namespace: "default",
						Annotations: map[string]string{
							"analysisDate": "2022-04-01",
						},
						Labels: map[string]string{
							"app": "example-app",
						},
					},
					Spec: batchv1.CronJobSpec{
						Schedule:          "*/1 * * * *",
						ConcurrencyPolicy: "Allow",
						JobTemplate: batchv1.JobTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									"app": "example-app",
								},
							},
							Spec: batchv1.JobSpec{
								Template: v1.PodTemplateSpec{
									Spec: v1.PodSpec{
										Containers: []v1.Container{
											{
												Name:  "example-container",
												Image: "nginx",
											},
										},
										RestartPolicy: v1.RestartPolicyOnFailure,
									},
								},
							},
						},
					},
				},
			),
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	cjAnalyzer := CronJobAnalyzer{}
	results, err := cjAnalyzer.Analyze(config)
	require.NoError(t, err)

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	expectations := []string{
		"default/CJ2",
		"default/CJ3",
		"default/CJ4",
	}

	require.Equal(t, len(expectations), len(results))

	for i, result := range results {
		require.Equal(t, expectations[i], result.Name)
	}
}
