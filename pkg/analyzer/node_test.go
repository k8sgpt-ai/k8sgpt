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

func TestNodeAnalyzerLabelSelectorFiltering(t *testing.T) {
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: fake.NewSimpleClientset(&v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "Node1",
					Namespace: "default",
					Labels: map[string]string{
						"app": "node",
					},
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Status: v1.ConditionFalse,
						},
					},
				},
			},
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Node2",
						Namespace: "default",
					},
					Status: v1.NodeStatus{
						Conditions: []v1.NodeCondition{
							{
								Type:   v1.NodeReady,
								Status: v1.ConditionFalse,
							},
						},
					},
				},
			),
		},
		Context:       context.Background(),
		Namespace:     "default",
		LabelSelector: "app=node",
	}

	nAnalyzer := NodeAnalyzer{}
	results, err := nAnalyzer.Analyze(config)
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, "Node1", results[0].Name)
}

func TestIsEKSNodeMonitoringAgentConditionType(t *testing.T) {
	tests := []struct {
		name          string
		conditionType v1.NodeConditionType
		expected      bool
	}{
		{
			name:          "AcceleratedHardwareReady is set by the EKS node monitoring agent",
			conditionType: "AcceleratedHardwareReady",
			expected:      true,
		},
		{
			name:          "ContainerRuntimeReady is set by the EKS node monitoring agent",
			conditionType: "ContainerRuntimeReady",
			expected:      true,
		},
		{
			name:          "KernelReady is set by the EKS node monitoring agent",
			conditionType: "KernelReady",
			expected:      true,
		},
		{
			name:          "NetworkingReady is set by the EKS node monitoring agent",
			conditionType: "NetworkingReady",
			expected:      true,
		},
		{
			name:          "StorageReady is set by the EKS node monitoring agent",
			conditionType: "StorageReady",
			expected:      true,
		},
		{
			name:          "Ready is a standard condition",
			conditionType: v1.NodeReady,
			expected:      false,
		},
		{
			name:          "MemoryPressure is a standard condition",
			conditionType: v1.NodeMemoryPressure,
			expected:      false,
		},
		{
			name:          "KernelDeadlock is not set by the EKS node monitoring agent",
			conditionType: "KernelDeadlock",
			expected:      false,
		},
		{
			name:          "empty type is not set by the EKS node monitoring agent",
			conditionType: "",
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, isEKSNodeMonitoringAgentConditionType(tt.conditionType))
		})
	}
}

func TestNodeAnalyzerEKSNodeMonitoringAgentConditionsTrueIgnored(t *testing.T) {
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: fake.NewSimpleClientset(&v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "Node1",
				},
				Status: v1.NodeStatus{
					// A healthy EKS node: the standard conditions plus the five
					// conditions the EKS node monitoring agent sets, all True.
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Status: v1.ConditionTrue,
						},
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
							Type:    "AcceleratedHardwareReady",
							Status:  v1.ConditionTrue,
							Reason:  "AcceleratedHardwareIsReady",
							Message: "Monitoring for the AcceleratedHardware system is active",
						},
						{
							Type:    "ContainerRuntimeReady",
							Status:  v1.ConditionTrue,
							Reason:  "ContainerRuntimeIsReady",
							Message: "Monitoring for the ContainerRuntime system is active",
						},
						{
							Type:    "KernelReady",
							Status:  v1.ConditionTrue,
							Reason:  "KernelIsReady",
							Message: "Monitoring for the Kernel system is active",
						},
						{
							Type:    "NetworkingReady",
							Status:  v1.ConditionTrue,
							Reason:  "NetworkingIsReady",
							Message: "Monitoring for the Networking system is active",
						},
						{
							Type:    "StorageReady",
							Status:  v1.ConditionTrue,
							Reason:  "DiskIsReady",
							Message: "Monitoring for the Disk system is active",
						},
					},
				},
			}),
		},
		Context: context.Background(),
	}

	nAnalyzer := NodeAnalyzer{}
	results, err := nAnalyzer.Analyze(config)
	require.NoError(t, err)
	require.Empty(t, results)
}

func TestNodeAnalyzerEKSNodeMonitoringAgentConditionFalseReported(t *testing.T) {
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: fake.NewSimpleClientset(&v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "Node1",
				},
				Status: v1.NodeStatus{
					Conditions: []v1.NodeCondition{
						{
							Type:   v1.NodeReady,
							Status: v1.ConditionTrue,
						},
						{
							Type:    "KernelReady",
							Status:  v1.ConditionTrue,
							Reason:  "KernelIsReady",
							Message: "Monitoring for the Kernel system is active",
						},
						// The agent reports a detected problem with Status False.
						{
							Type:    "NetworkingReady",
							Status:  v1.ConditionFalse,
							Reason:  "IPAMDNotReady",
							Message: "IPAM-D has failed to connect to API Server",
						},
					},
				},
			}),
		},
		Context: context.Background(),
	}

	nAnalyzer := NodeAnalyzer{}
	results, err := nAnalyzer.Analyze(config)
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, "Node1", results[0].Name)
	require.Equal(t, 1, len(results[0].Error))
	require.Equal(t, "Node1 has condition of type NetworkingReady, reason IPAMDNotReady: IPAM-D has failed to connect to API Server", results[0].Error[0].Text)
}
