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

func TestPodAnalyzer(t *testing.T) {
	tests := []struct {
		name         string
		config       common.Analyzer
		expectations []struct {
			name          string
			failuresCount int
		}
	}{
		{
			name: "Pending pods, namespace filtering and readiness probe failure",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "Pod1",
								Namespace: "default",
							},
							Status: v1.PodStatus{
								Phase: v1.PodPending,
								Conditions: []v1.PodCondition{
									{
										// This condition will contribute to failures.
										Type:    v1.PodScheduled,
										Reason:  "Unschedulable",
										Message: "0/1 nodes are available: 1 node(s) had taint {node-role.kubernetes.io/master: }, that the pod didn't tolerate.",
									},
									{
										// This condition won't contribute to failures.
										Type:   v1.PodScheduled,
										Reason: "Unexpected failure",
									},
								},
							},
						},
						&v1.Pod{
							// This pod won't be selected because of namespace filtering.
							ObjectMeta: metav1.ObjectMeta{
								Name:      "Pod2",
								Namespace: "test",
							},
						},
						&v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "Pod3",
								Namespace: "default",
							},
							Status: v1.PodStatus{
								// When pod is Running but its ReadinessProbe fails
								Phase: v1.PodRunning,
								ContainerStatuses: []v1.ContainerStatus{
									{
										Ready: false,
									},
								},
							},
						},
						&v1.Event{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "Event1",
								Namespace: "default",
							},
							InvolvedObject: v1.ObjectReference{
								Kind:      "Pod",
								Name:      "Pod3",
								Namespace: "default",
							},
							Reason:  "Unhealthy",
							Message: "readiness probe failed: the detail reason here ...",
							Source:  v1.EventSource{Component: "eventTest"},
							Count:   1,
							Type:    v1.EventTypeWarning,
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
					name:          "default/Pod1",
					failuresCount: 1,
				},
				{
					name:          "default/Pod3",
					failuresCount: 1,
				},
			},
		},
		{
			name: "readiness probe failure without any event",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "Pod1",
								Namespace: "default",
							},
							Status: v1.PodStatus{
								// When pod is Running but its ReadinessProbe fails
								// It won't contribute to any failures because
								// there's no event present.
								Phase: v1.PodRunning,
								ContainerStatuses: []v1.ContainerStatus{
									{
										Ready: false,
									},
								},
							},
						},
					),
				},
				Context:   context.Background(),
				Namespace: "default",
			},
		},
		{
			name: "Init container status state waiting",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "Pod1",
								Namespace: "default",
							},
							Status: v1.PodStatus{
								Phase: v1.PodPending,
								InitContainerStatuses: []v1.ContainerStatus{
									{
										Ready: true,
										State: v1.ContainerState{
											Running: &v1.ContainerStateRunning{
												StartedAt: metav1.Now(),
											},
										},
									},
									{
										Ready: false,
										State: v1.ContainerState{
											Waiting: &v1.ContainerStateWaiting{
												// This represents a container that is still being created or blocked due to conditions such as OOMKilled
												Reason: "ContainerCreating",
											},
										},
									},
								},
							},
						},
						&v1.Event{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "Event1",
								Namespace: "default",
							},
							InvolvedObject: v1.ObjectReference{
								Kind:      "Pod",
								Name:      "Pod1",
								Namespace: "default",
							},
							Reason:  "FailedCreatePodSandBox",
							Message: "failed to create the pod sandbox ...",
							Type:    v1.EventTypeWarning,
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
					name:          "default/Pod1",
					failuresCount: 1,
				},
			},
		},
		{
			name: "Container status state waiting but no event reported",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "Pod1",
								Namespace: "default",
							},
							Status: v1.PodStatus{
								Phase: v1.PodPending,
								ContainerStatuses: []v1.ContainerStatus{
									{
										Ready: false,
										State: v1.ContainerState{
											Waiting: &v1.ContainerStateWaiting{
												// This represents a container that is still being created or blocked due to conditions such as OOMKilled
												Reason: "ContainerCreating",
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
			},
		},
		{
			name: "Container status state waiting",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&v1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "Pod1",
								Namespace: "default",
							},
							Status: v1.PodStatus{
								Phase: v1.PodPending,
								ContainerStatuses: []v1.ContainerStatus{
									{
										Name:  "Container1",
										Ready: false,
										State: v1.ContainerState{
											Waiting: &v1.ContainerStateWaiting{
												// This represents a container that is still being created or blocked due to conditions such as OOMKilled
												Reason: "ContainerCreating",
											},
										},
									},
									{
										Name:  "Container2",
										Ready: false,
										State: v1.ContainerState{
											Waiting: &v1.ContainerStateWaiting{
												// This represents container that is in CrashLoopBackOff state due to conditions such as OOMKilled
												Reason: "CrashLoopBackOff",
											},
										},
										LastTerminationState: v1.ContainerState{
											Terminated: &v1.ContainerStateTerminated{
												Reason: "test reason",
											},
										},
									},
									{
										Name:  "Container3",
										Ready: false,
										State: v1.ContainerState{
											Waiting: &v1.ContainerStateWaiting{
												// This won't contribute to failures.
												Reason:  "RandomReason",
												Message: "This container won't be present in the failures",
											},
										},
									},
									{
										Name:  "Container4",
										Ready: false,
										State: v1.ContainerState{
											Waiting: &v1.ContainerStateWaiting{
												// Valid error reason.
												Reason:  "PreStartHookError",
												Message: "Container4 encountered PreStartHookError",
											},
										},
									},
									{
										Name:  "Container5",
										Ready: false,
										State: v1.ContainerState{
											Waiting: &v1.ContainerStateWaiting{
												// Valid error reason.
												Reason:  "CrashLoopBackOff",
												Message: "Container4 encountered CrashLoopBackOff",
											},
										},
									},
								},
							},
						},
						&v1.Event{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "Event1",
								Namespace: "default",
							},
							InvolvedObject: v1.ObjectReference{
								Kind:      "Pod",
								Name:      "Pod1",
								Namespace: "default",
							},
							// This reason won't contribute to failures.
							Reason: "RandomEvent",
							Type:   v1.EventTypeWarning,
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
					name:          "default/Pod1",
					failuresCount: 3,
				},
			},
		},
	}

	podAnalyzer := PodAnalyzer{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := podAnalyzer.Analyze(tt.config)
			require.NoError(t, err)

			if tt.expectations == nil {
				require.Equal(t, 0, len(results))
			} else {
				sort.Slice(results, func(i, j int) bool {
					return results[i].Name < results[j].Name
				})

				require.Equal(t, len(tt.expectations), len(results))

				for i, result := range results {
					require.Equal(t, tt.expectations[i].name, result.Name)
					require.Equal(t, tt.expectations[i].failuresCount, len(result.Error))
				}
			}
		})
	}
}
