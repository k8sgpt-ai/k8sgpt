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

func TestHealthyNodeConditionStatus(t *testing.T) {
	tests := []struct {
		name          string
		conditionType v1.NodeConditionType
		healthy       v1.ConditionStatus
		known         bool
	}{
		{
			name:          "Ready is healthy when True",
			conditionType: v1.NodeReady,
			healthy:       v1.ConditionTrue,
			known:         true,
		},
		{
			name:          "AcceleratedHardwareReady from the EKS node monitoring agent is healthy when True",
			conditionType: "AcceleratedHardwareReady",
			healthy:       v1.ConditionTrue,
			known:         true,
		},
		{
			name:          "ContainerRuntimeReady from the EKS node monitoring agent is healthy when True",
			conditionType: "ContainerRuntimeReady",
			healthy:       v1.ConditionTrue,
			known:         true,
		},
		{
			name:          "KernelReady from the EKS node monitoring agent is healthy when True",
			conditionType: "KernelReady",
			healthy:       v1.ConditionTrue,
			known:         true,
		},
		{
			name:          "NetworkingReady from the EKS node monitoring agent is healthy when True",
			conditionType: "NetworkingReady",
			healthy:       v1.ConditionTrue,
			known:         true,
		},
		{
			name:          "StorageReady from the EKS node monitoring agent is healthy when True",
			conditionType: "StorageReady",
			healthy:       v1.ConditionTrue,
			known:         true,
		},
		{
			name:          "MemoryPressure is healthy when False",
			conditionType: v1.NodeMemoryPressure,
			healthy:       v1.ConditionFalse,
			known:         true,
		},
		{
			name:          "DiskPressure is healthy when False",
			conditionType: v1.NodeDiskPressure,
			healthy:       v1.ConditionFalse,
			known:         true,
		},
		{
			name:          "PIDPressure is healthy when False",
			conditionType: v1.NodePIDPressure,
			healthy:       v1.ConditionFalse,
			known:         true,
		},
		{
			name:          "NetworkUnavailable is healthy when False",
			conditionType: v1.NodeNetworkUnavailable,
			healthy:       v1.ConditionFalse,
			known:         true,
		},
		{
			name:          "ContainerRuntimeUnhealthy from node-problem-detector is healthy when False",
			conditionType: "ContainerRuntimeUnhealthy",
			healthy:       v1.ConditionFalse,
			known:         true,
		},
		{
			name:          "CorruptDockerOverlay2 from node-problem-detector is healthy when False",
			conditionType: "CorruptDockerOverlay2",
			healthy:       v1.ConditionFalse,
			known:         true,
		},
		{
			name:          "CperHardwareErrorFatal from node-problem-detector is healthy when False",
			conditionType: "CperHardwareErrorFatal",
			healthy:       v1.ConditionFalse,
			known:         true,
		},
		{
			name:          "DiskBadBlock from node-problem-detector is healthy when False",
			conditionType: "DiskBadBlock",
			healthy:       v1.ConditionFalse,
			known:         true,
		},
		{
			name:          "FrequentContainerdRestart from node-problem-detector is healthy when False",
			conditionType: "FrequentContainerdRestart",
			healthy:       v1.ConditionFalse,
			known:         true,
		},
		{
			name:          "FrequentDockerRestart from node-problem-detector is healthy when False",
			conditionType: "FrequentDockerRestart",
			healthy:       v1.ConditionFalse,
			known:         true,
		},
		{
			name:          "FrequentKubeletRestart from node-problem-detector is healthy when False",
			conditionType: "FrequentKubeletRestart",
			healthy:       v1.ConditionFalse,
			known:         true,
		},
		{
			name:          "FrequentUnregisterNetDevice from node-problem-detector is healthy when False",
			conditionType: "FrequentUnregisterNetDevice",
			healthy:       v1.ConditionFalse,
			known:         true,
		},
		{
			name:          "KernelDeadlock from node-problem-detector is healthy when False",
			conditionType: "KernelDeadlock",
			healthy:       v1.ConditionFalse,
			known:         true,
		},
		{
			name:          "KubeProxyUnhealthy from node-problem-detector is healthy when False",
			conditionType: "KubeProxyUnhealthy",
			healthy:       v1.ConditionFalse,
			known:         true,
		},
		{
			name:          "KubeletUnhealthy from node-problem-detector is healthy when False",
			conditionType: "KubeletUnhealthy",
			healthy:       v1.ConditionFalse,
			known:         true,
		},
		{
			name:          "NTPProblem from node-problem-detector is healthy when False",
			conditionType: "NTPProblem",
			healthy:       v1.ConditionFalse,
			known:         true,
		},
		{
			name:          "ReadonlyFilesystem from node-problem-detector is healthy when False",
			conditionType: "ReadonlyFilesystem",
			healthy:       v1.ConditionFalse,
			known:         true,
		},
		{
			name:          "XfsShutdown from node-problem-detector is healthy when False",
			conditionType: "XfsShutdown",
			healthy:       v1.ConditionFalse,
			known:         true,
		},
		{
			name:          "KubeletProblem is not known",
			conditionType: "KubeletProblem",
			healthy:       "",
			known:         false,
		},
		{
			name:          "empty type is not known",
			conditionType: "",
			healthy:       "",
			known:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			healthy, known := healthyNodeConditionStatus(tt.conditionType)
			require.Equal(t, tt.known, known)
			require.Equal(t, tt.healthy, healthy)
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

func TestNodeAnalyzerNodeProblemDetectorConditionsFalseIgnored(t *testing.T) {
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: fake.NewSimpleClientset(&v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "Node1",
				},
				Status: v1.NodeStatus{
					// A healthy node running node-problem-detector: the standard
					// conditions plus the conditions from the configs in its
					// repository, all False with their default reason and message.
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
							Type:    "ContainerRuntimeUnhealthy",
							Status:  v1.ConditionFalse,
							Reason:  "ContainerRuntimeIsHealthy",
							Message: "Container runtime on the node is functioning properly",
						},
						{
							Type:    "CorruptDockerOverlay2",
							Status:  v1.ConditionFalse,
							Reason:  "NoCorruptDockerOverlay2",
							Message: "docker overlay2 is functioning properly",
						},
						{
							Type:    "CperHardwareErrorFatal",
							Status:  v1.ConditionFalse,
							Reason:  "CperHardwareHasNoFatalError",
							Message: "UEFI CPER has no fatal error",
						},
						{
							Type:    "DiskBadBlock",
							Status:  v1.ConditionFalse,
							Reason:  "DiskBadBlock",
							Message: "Disk no bad block",
						},
						{
							Type:    "FrequentContainerdRestart",
							Status:  v1.ConditionFalse,
							Reason:  "NoFrequentContainerdRestart",
							Message: "containerd is functioning properly",
						},
						{
							Type:    "FrequentDockerRestart",
							Status:  v1.ConditionFalse,
							Reason:  "NoFrequentDockerRestart",
							Message: "docker is functioning properly",
						},
						{
							Type:    "FrequentKubeletRestart",
							Status:  v1.ConditionFalse,
							Reason:  "NoFrequentKubeletRestart",
							Message: "kubelet is functioning properly",
						},
						{
							Type:    "FrequentUnregisterNetDevice",
							Status:  v1.ConditionFalse,
							Reason:  "NoFrequentUnregisterNetDevice",
							Message: "node is functioning properly",
						},
						{
							Type:    "KernelDeadlock",
							Status:  v1.ConditionFalse,
							Reason:  "KernelHasNoDeadlock",
							Message: "kernel has no deadlock",
						},
						{
							Type:    "KubeProxyUnhealthy",
							Status:  v1.ConditionFalse,
							Reason:  "KubeProxyIsHealthy",
							Message: "kube-proxy on the node is functioning properly",
						},
						{
							Type:    "KubeletUnhealthy",
							Status:  v1.ConditionFalse,
							Reason:  "KubeletIsHealthy",
							Message: "kubelet on the node is functioning properly",
						},
						{
							Type:    "NTPProblem",
							Status:  v1.ConditionFalse,
							Reason:  "NTPIsUp",
							Message: "ntp service is up",
						},
						{
							Type:    "ReadonlyFilesystem",
							Status:  v1.ConditionFalse,
							Reason:  "FilesystemIsNotReadOnly",
							Message: "Filesystem is not read-only",
						},
						{
							Type:    "XfsShutdown",
							Status:  v1.ConditionFalse,
							Reason:  "XfsHasNotShutDown",
							Message: "XFS has not shutdown",
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

func TestNodeAnalyzerNodeProblemDetectorConditionTrueReported(t *testing.T) {
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
							Type:    "ReadonlyFilesystem",
							Status:  v1.ConditionFalse,
							Reason:  "FilesystemIsNotReadOnly",
							Message: "Filesystem is not read-only",
						},
						// node-problem-detector reports a detected problem with Status True.
						{
							Type:    "KernelDeadlock",
							Status:  v1.ConditionTrue,
							Reason:  "DockerHung",
							Message: "INFO: task docker:20744 blocked for more than 120 seconds.",
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
	require.Equal(t, "Node1 has condition of type KernelDeadlock, reason DockerHung: INFO: task docker:20744 blocked for more than 120 seconds.", results[0].Error[0].Text)
}

func TestNodeAnalyzerNodeProblemDetectorConditionUnknownReported(t *testing.T) {
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
						// A custom plugin that cannot run reports Status Unknown.
						{
							Type:    "KubeletUnhealthy",
							Status:  v1.ConditionUnknown,
							Reason:  "KubeletIsHealthy",
							Message: "Error in starting plugin. Please check the error log",
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
	require.Equal(t, "Node1 has condition of type KubeletUnhealthy, reason KubeletIsHealthy: Error in starting plugin. Please check the error log", results[0].Error[0].Text)
}

func TestNodeAnalyzerEtcdIsVoterIgnored(t *testing.T) {
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
						// k3s sets EtcdIsVoter on control-plane nodes. Both True and
						// False are normal.
						{
							Type:    "EtcdIsVoter",
							Status:  v1.ConditionTrue,
							Reason:  "MemberNotLearner",
							Message: "Node is a voting member of the etcd cluster",
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

func TestNodeAnalyzerStandardConditionNewStatusReported(t *testing.T) {
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
						// Only False is healthy for NetworkUnavailable, so a status
						// other than True, False and Unknown is reported as well.
						{
							Type:    v1.NodeNetworkUnavailable,
							Status:  "CompletelyUnknown",
							Reason:  "NetworkStatus",
							Message: "Network status is neither True, False nor Unknown",
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
	require.Equal(t, "Node1 has condition of type NetworkUnavailable, reason NetworkStatus: Network status is neither True, False nor Unknown", results[0].Error[0].Text)
}
