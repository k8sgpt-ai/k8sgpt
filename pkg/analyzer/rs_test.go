/*
Copyright 2024 The K8sGPT Authors.
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
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestReplicaSetAnalyzer(t *testing.T) {
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: fake.NewSimpleClientset(
				&appsv1.ReplicaSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ReplicaSet1",
						Namespace: "default",
					},
					Status: appsv1.ReplicaSetStatus{
						Replicas: 0,
						Conditions: []appsv1.ReplicaSetCondition{
							{
								// Should contribute to failures.
								Type:    appsv1.ReplicaSetReplicaFailure,
								Reason:  "FailedCreate",
								Message: "failed to create test replica set 1",
							},
						},
					},
				},
				&appsv1.ReplicaSet{
					// This replicaset won't be discovered as it is not in the
					// default namespace.
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ReplicaSet2",
						Namespace: "test",
					},
				},
				&appsv1.ReplicaSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ReplicaSet3",
						Namespace: "default",
					},
					Status: appsv1.ReplicaSetStatus{
						Replicas: 0,
						Conditions: []appsv1.ReplicaSetCondition{
							{
								Type: appsv1.ReplicaSetReplicaFailure,
								// Should not be included in the failures.
								Reason: "RandomError",
							},
						},
					},
				},
				&appsv1.ReplicaSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ReplicaSet4",
						Namespace: "default",
					},
					Status: appsv1.ReplicaSetStatus{
						Replicas: 0,
						Conditions: []appsv1.ReplicaSetCondition{
							{
								// Should contribute to failures.
								Type:    appsv1.ReplicaSetReplicaFailure,
								Reason:  "FailedCreate",
								Message: "failed to create test replica set 4 condition 1",
							},
							{
								// Should not contribute to failures.
								Type:   appsv1.ReplicaSetReplicaFailure,
								Reason: "Unknown",
							},
							{
								// Should not contribute to failures.
								Type:    appsv1.ReplicaSetReplicaFailure,
								Reason:  "FailedCreate",
								Message: "failed to create test replica set 4 condition 3",
							},
						},
					},
				},
				&appsv1.ReplicaSet{
					// Replicaset without any failures.
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ReplicaSet5",
						Namespace: "default",
					},
					Status: appsv1.ReplicaSetStatus{
						Replicas: 3,
					},
				},
			),
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	rsAnalyzer := ReplicaSetAnalyzer{}
	results, err := rsAnalyzer.Analyze(config)
	require.NoError(t, err)

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	expectations := []struct {
		name          string
		failuresCount int
	}{
		{
			name:          "default/ReplicaSet1",
			failuresCount: 1,
		},
		{
			name:          "default/ReplicaSet4",
			failuresCount: 2,
		},
	}

	require.Equal(t, len(expectations), len(results))

	for i, result := range results {
		require.Equal(t, expectations[i].name, result.Name)
		require.Equal(t, expectations[i].failuresCount, len(result.Error))
	}
}
