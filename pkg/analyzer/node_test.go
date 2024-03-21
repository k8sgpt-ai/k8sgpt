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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNodeAnalyzer(t *testing.T) {
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: fake.NewSimpleClientset(
				&v1.Node{
					// A node without Status Conditions shouldn't contribute to failures.
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Node1",
						Namespace: "test",
					},
				},
				&v1.Node{
					// Nodes are not filtered using namespace.
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Node2",
						Namespace: "default",
					},
					Status: v1.NodeStatus{
						Conditions: []v1.NodeCondition{
							{
								// Won't contribute to failures.
								Type:   v1.NodeReady,
								Status: v1.ConditionTrue,
							},
							{
								// Will contribute to failures.
								Type:   v1.NodeReady,
								Status: v1.ConditionFalse,
							},
							{
								// Will contribute to failures.
								Type:   v1.NodeReady,
								Status: v1.ConditionUnknown,
							},
							// Non-false statuses for the default cases contribute to failures.
							{
								Type:   v1.NodeMemoryPressure,
								Status: v1.ConditionTrue,
							},
							{
								Type:   v1.NodeDiskPressure,
								Status: v1.ConditionTrue,
							},
							{
								Type:   v1.NodePIDPressure,
								Status: v1.ConditionTrue,
							},
							{
								Type:   v1.NodeNetworkUnavailable,
								Status: v1.ConditionTrue,
							},
							{
								Type:   v1.NodeMemoryPressure,
								Status: v1.ConditionUnknown,
							},
							{
								Type:   v1.NodeDiskPressure,
								Status: v1.ConditionUnknown,
							},
							{
								Type:   v1.NodePIDPressure,
								Status: v1.ConditionUnknown,
							},
							{
								Type:   v1.NodeNetworkUnavailable,
								Status: v1.ConditionUnknown,
							},
							// A cloud provider may set their own condition and/or a new status
							// might be introduced. In such cases a failure is assumed and
							// the code shouldn't break, although it might be a false positive.
							{
								Type:   "UnknownNodeConditionType",
								Status: "CompletelyUnknown",
							},
							// These won't contribute to failures.
							{
								Type:   v1.NodeMemoryPressure,
								Status: v1.ConditionFalse,
							},
							{
								Type:   v1.NodeDiskPressure,
								Status: v1.ConditionFalse,
							},
							{
								Type:   v1.NodePIDPressure,
								Status: v1.ConditionFalse,
							},
							{
								Type:   v1.NodeNetworkUnavailable,
								Status: v1.ConditionFalse,
							},
						},
					},
				},
				&v1.Node{
					// A node without any failures shouldn't be present in the results.
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Node3",
						Namespace: "test",
					},
					Status: v1.NodeStatus{
						Conditions: []v1.NodeCondition{
							{
								// Won't contribute to failures.
								Type:   v1.NodeReady,
								Status: v1.ConditionTrue,
							},
						},
					},
				},
			),
		},
		Context:   context.Background(),
		Namespace: "test",
	}

	nAnalyzer := NodeAnalyzer{}
	results, err := nAnalyzer.Analyze(config)
	require.NoError(t, err)

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	expectations := []struct {
		name          string
		failuresCount int
	}{
		{
			name:          "Node2",
			failuresCount: 11,
		},
	}

	require.Equal(t, len(expectations), len(results))

	for i, result := range results {
		require.Equal(t, expectations[i].name, result.Name)
		require.Equal(t, expectations[i].failuresCount, len(result.Error))
	}
}
