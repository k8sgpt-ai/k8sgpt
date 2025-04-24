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

func TestServiceAnalyzer(t *testing.T) {
	tests := []struct {
		name         string
		config       common.Analyzer
		expectations []struct {
			name          string
			failuresCount int
		}
	}{
		{
			name: "Service with no endpoints",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&v1.Endpoints{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-service",
								Namespace: "default",
							},
							Subsets: []v1.EndpointSubset{}, // Empty subsets
						},
						&v1.Service{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-service",
								Namespace: "default",
							},
							Spec: v1.ServiceSpec{
								Selector: map[string]string{
									"app": "test",
								},
							},
						},
					),
				},
				Namespace: "default",
			},
			expectations: []struct {
				name          string
				failuresCount int
			}{
				{
					name:          "default/test-service",
					failuresCount: 1, // One failure for no endpoints
				},
			},
		},
		{
			name: "Service with not ready endpoints",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&v1.Endpoints{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-service",
								Namespace: "default",
							},
							Subsets: []v1.EndpointSubset{
								{
									NotReadyAddresses: []v1.EndpointAddress{
										{
											TargetRef: &v1.ObjectReference{
												Kind: "Pod",
												Name: "test-pod",
											},
										},
									},
								},
							},
						},
						&v1.Service{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-service",
								Namespace: "default",
							},
							Spec: v1.ServiceSpec{
								Selector: map[string]string{
									"app": "test",
								},
							},
						},
					),
				},
				Namespace: "default",
			},
			expectations: []struct {
				name          string
				failuresCount int
			}{
				{
					name:          "default/test-service",
					failuresCount: 1, // One failure for not ready endpoints
				},
			},
		},
		{
			name: "Service with warning events",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&v1.Endpoints{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-service",
								Namespace: "default",
							},
							Subsets: []v1.EndpointSubset{}, // Empty subsets
						},
						&v1.Service{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-service",
								Namespace: "default",
							},
							Spec: v1.ServiceSpec{
								Selector: map[string]string{
									"app": "test",
								},
							},
						},
						&v1.Event{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-event",
								Namespace: "default",
							},
							InvolvedObject: v1.ObjectReference{
								Kind:      "Service",
								Name:      "test-service",
								Namespace: "default",
							},
							Type:    "Warning",
							Reason:  "TestReason",
							Message: "Test warning message",
						},
					),
				},
				Namespace: "default",
			},
			expectations: []struct {
				name          string
				failuresCount int
			}{
				{
					name:          "default/test-service",
					failuresCount: 2, // One failure for no endpoints, one for warning event
				},
			},
		},
		{
			name: "Service with leader election annotation",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&v1.Endpoints{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-service",
								Namespace: "default",
								Annotations: map[string]string{
									"control-plane.alpha.kubernetes.io/leader": "test-leader",
								},
							},
							Subsets: []v1.EndpointSubset{}, // Empty subsets
						},
						&v1.Service{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-service",
								Namespace: "default",
							},
							Spec: v1.ServiceSpec{
								Selector: map[string]string{
									"app": "test",
								},
							},
						},
					),
				},
				Namespace: "default",
			},
			expectations: []struct {
				name          string
				failuresCount int
			}{
				// No expectations for leader election endpoints
			},
		},
		{
			name: "Service with non-existent service",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&v1.Endpoints{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-service",
								Namespace: "default",
							},
							Subsets: []v1.EndpointSubset{}, // Empty subsets
						},
					),
				},
				Namespace: "default",
			},
			expectations: []struct {
				name          string
				failuresCount int
			}{
				// No expectations for non-existent service
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := ServiceAnalyzer{}
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

func TestServiceAnalyzerLabelSelectorFiltering(t *testing.T) {
	clientSet :=
		fake.NewSimpleClientset(
			&v1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "Endpoint1",
					Namespace: "default",
					Labels: map[string]string{
						"app":     "service",
						"part-of": "test",
					},
				},
				// Endpoint with non-zero subsets.
				Subsets: []v1.EndpointSubset{
					{
						// These not ready end points will contribute to failures.
						NotReadyAddresses: []v1.EndpointAddress{
							{
								TargetRef: &v1.ObjectReference{
									Kind: "test-reference",
									Name: "reference1",
								},
							},
							{
								TargetRef: &v1.ObjectReference{
									Kind: "test-reference",
									Name: "reference2",
								},
							},
						},
					},
					{
						// These not ready end points will contribute to failures.
						NotReadyAddresses: []v1.EndpointAddress{
							{
								TargetRef: &v1.ObjectReference{
									Kind: "test-reference",
									Name: "reference3",
								},
							},
						},
					},
				},
			},
			&v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "Service1",
					Namespace: "default",
					Labels: map[string]string{
						"app": "service",
					},
				},
				Spec: v1.ServiceSpec{
					Selector: map[string]string{
						"app1": "test-app1",
						"app2": "test-app2",
					},
				},
			},
			&v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "Service2",
					Namespace: "default",
				},
				Spec: v1.ServiceSpec{
					Selector: map[string]string{
						"app1": "test-app1",
						"app2": "test-app2",
					},
				},
			},
		)
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientSet,
		},
		Context:       context.Background(),
		Namespace:     "default",
		LabelSelector: "app=service",
	}

	sAnalyzer := ServiceAnalyzer{}
	results, err := sAnalyzer.Analyze(config)
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, "default/Endpoint1", results[0].Name)

	config = common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientSet,
		},
		Context:       context.Background(),
		Namespace:     "default",
		LabelSelector: "app=service,part-of=test",
	}

	sAnalyzer = ServiceAnalyzer{}
	results, err = sAnalyzer.Analyze(config)
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, "default/Endpoint1", results[0].Name)
}
