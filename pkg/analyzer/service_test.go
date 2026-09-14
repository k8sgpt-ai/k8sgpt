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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
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
			name: "Service with not ready endpoints with nil TargetRef",
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
									// Not-ready address with no TargetRef (bare-IP /
									// manually-created Endpoints, not backed by a Pod).
									NotReadyAddresses: []v1.EndpointAddress{
										{
											IP: "10.0.0.1",
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

func TestServiceAnalyzer_EventKindFiltering(t *testing.T) {
	// Create a clientset with reactor to properly emulate API server field selector behavior
	cs := fake.NewSimpleClientset()

	// Create test objects
	testEndpoint := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ollama",
			Namespace: "k8sgpt",
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP: "10.0.0.1",
						TargetRef: &v1.ObjectReference{
							Kind: "Pod",
							Name: "ollama-pod",
						},
					},
				},
			},
		},
	}

	testService := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ollama",
			Namespace: "k8sgpt",
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": "k8sgpt",
			},
		},
	}

	// Event on the K8sGPT CR (wrong kind - should NOT be attributed to Service)
	k8sgptCREvent := &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "k8sgpt-cr-event",
			Namespace: "k8sgpt",
		},
		InvolvedObject: v1.ObjectReference{
			Kind:      "K8sGPT",
			Name:      "ollama",
			Namespace: "k8sgpt",
		},
		Type:    "Warning",
		Reason:  "AnalysisFailed",
		Message: "failed to call Analyze RPC: rpc error: code = Unavailable desc = error reading from server: EOF",
	}

	// Event on the Service itself (correct kind - SHOULD be attributed)
	serviceEvent := &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service-event",
			Namespace: "k8sgpt",
		},
		InvolvedObject: v1.ObjectReference{
			Kind:      "Service",
			Name:      "ollama",
			Namespace: "k8sgpt",
		},
		Type:    "Warning",
		Reason:  "ServiceIssue",
		Message: "actual service warning",
	}

	_, err := cs.CoreV1().Endpoints("k8sgpt").Create(context.Background(), testEndpoint, metav1.CreateOptions{})
	require.NoError(t, err)
	_, err = cs.CoreV1().Services("k8sgpt").Create(context.Background(), testService, metav1.CreateOptions{})
	require.NoError(t, err)

	// Capture the field selector sent to the API
	capturedFieldSelector := ""

	// Add reactor to emulate API server field selector behavior
	// The fake clientset does NOT filter by field selector, so we must do it manually
	cs.PrependReactor("list", "events", func(action k8stesting.Action) (bool, runtime.Object, error) {
		listAction, ok := action.(k8stesting.ListAction)
		if !ok {
			return false, nil, nil
		}

		fieldSel := listAction.GetListRestrictions().Fields
		capturedFieldSelector = fieldSel.String()

		// Manually filter events based on field selector
		allEvents := &v1.EventList{}
		allEvents.Items = append(allEvents.Items, *k8sgptCREvent, *serviceEvent)

		filteredEvents := &v1.EventList{}
		for _, event := range allEvents.Items {
			match := true

			// Check involvedObject.kind
			if kindVal, found := fieldSel.RequiresExactMatch("involvedObject.kind"); found {
				if event.InvolvedObject.Kind != kindVal {
					match = false
				}
			}

			// Check involvedObject.name
			if nameVal, found := fieldSel.RequiresExactMatch("involvedObject.name"); found {
				if event.InvolvedObject.Name != nameVal {
					match = false
				}
			}

			if match {
				filteredEvents.Items = append(filteredEvents.Items, event)
			}
		}

		return true, filteredEvents, nil
	})

	// Run the analyzer
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: cs,
		},
		Context:   context.Background(),
		Namespace: "k8sgpt",
	}

	analyzer := ServiceAnalyzer{}
	results, err := analyzer.Analyze(config)
	require.NoError(t, err)

	// Verify the field selector includes kind filter
	require.Contains(t, capturedFieldSelector, "involvedObject.kind=Service",
		"Field selector should filter by involvedObject.kind=Service")
	require.Contains(t, capturedFieldSelector, "involvedObject.name=ollama",
		"Field selector should filter by involvedObject.name")

	// Should have exactly 1 result (Service with 1 failure from the Service event only)
	require.Len(t, results, 1, "Should have 1 result for the Service")
	require.Equal(t, "k8sgpt/ollama", results[0].Name)

	// Should have exactly 1 failure (the Service event, not the K8sGPT CR event)
	require.Len(t, results[0].Error, 1, "Should have 1 failure from the Service event only")
	require.Contains(t, results[0].Error[0].Text, "actual service warning",
		"Failure should be from the Service event, not the K8sGPT CR event")
	require.NotContains(t, results[0].Error[0].Text, "failed to call Analyze RPC",
		"Should not include the K8sGPT CR event")
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
