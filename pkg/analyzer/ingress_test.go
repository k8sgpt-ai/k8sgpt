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
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestIngressAnalyzer(t *testing.T) {
	// Create test cases
	testCases := []struct {
		name           string
		ingress        *networkingv1.Ingress
		expectedIssues []string
	}{
		{
			name: "Non-existent backend service",
			ingress: &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ingress",
					Namespace: "default",
				},
				Spec: networkingv1.IngressSpec{
					Rules: []networkingv1.IngressRule{
						{
							Host: "example.com",
							IngressRuleValue: networkingv1.IngressRuleValue{
								HTTP: &networkingv1.HTTPIngressRuleValue{
									Paths: []networkingv1.HTTPIngressPath{
										{
											Path: "/",
											Backend: networkingv1.IngressBackend{
												Service: &networkingv1.IngressServiceBackend{
													Name: "non-existent-service",
													Port: networkingv1.ServiceBackendPort{
														Number: 80,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedIssues: []string{
				"Ingress default/test-ingress does not specify an Ingress class.",
				"Ingress uses the service default/non-existent-service which does not exist.",
			},
		},
		{
			name: "Non-existent TLS secret",
			ingress: &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ingress-tls",
					Namespace: "default",
				},
				Spec: networkingv1.IngressSpec{
					TLS: []networkingv1.IngressTLS{
						{
							Hosts:      []string{"example.com"},
							SecretName: "non-existent-secret",
						},
					},
					Rules: []networkingv1.IngressRule{
						{
							Host: "example.com",
							IngressRuleValue: networkingv1.IngressRuleValue{
								HTTP: &networkingv1.HTTPIngressRuleValue{
									Paths: []networkingv1.HTTPIngressPath{
										{
											Path: "/",
											Backend: networkingv1.IngressBackend{
												Service: &networkingv1.IngressServiceBackend{
													Name: "test-service",
													Port: networkingv1.ServiceBackendPort{
														Number: 80,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedIssues: []string{
				"Ingress default/test-ingress-tls does not specify an Ingress class.",
				"Ingress uses the service default/test-service which does not exist.",
				"Ingress uses the secret default/non-existent-secret as a TLS certificate which does not exist.",
			},
		},
		{
			name: "Multiple issues",
			ingress: &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ingress-multi",
					Namespace: "default",
				},
				Spec: networkingv1.IngressSpec{
					TLS: []networkingv1.IngressTLS{
						{
							Hosts:      []string{"example.com"},
							SecretName: "non-existent-secret",
						},
					},
					Rules: []networkingv1.IngressRule{
						{
							Host: "example.com",
							IngressRuleValue: networkingv1.IngressRuleValue{
								HTTP: &networkingv1.HTTPIngressRuleValue{
									Paths: []networkingv1.HTTPIngressPath{
										{
											Path: "/",
											Backend: networkingv1.IngressBackend{
												Service: &networkingv1.IngressServiceBackend{
													Name: "non-existent-service",
													Port: networkingv1.ServiceBackendPort{
														Number: 80,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedIssues: []string{
				"Ingress default/test-ingress-multi does not specify an Ingress class.",
				"Ingress uses the service default/non-existent-service which does not exist.",
				"Ingress uses the secret default/non-existent-secret as a TLS certificate which does not exist.",
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new context and clientset for each test case
			ctx := context.Background()
			clientset := fake.NewSimpleClientset()

			// Create the ingress in the fake clientset
			_, err := clientset.NetworkingV1().Ingresses(tc.ingress.Namespace).Create(ctx, tc.ingress, metav1.CreateOptions{})
			assert.NoError(t, err)

			// Create the analyzer configuration
			config := common.Analyzer{
				Client: &kubernetes.Client{
					Client: clientset,
				},
				Context:   ctx,
				Namespace: tc.ingress.Namespace,
			}

			// Create the analyzer and run analysis
			analyzer := IngressAnalyzer{}
			results, err := analyzer.Analyze(config)
			assert.NoError(t, err)

			// Check that we got the expected number of issues
			assert.Len(t, results, 1, "Expected 1 result")
			result := results[0]
			assert.Len(t, result.Error, len(tc.expectedIssues), "Expected %d issues, got %d", len(tc.expectedIssues), len(result.Error))

			// Check that each expected issue is present
			for _, expectedIssue := range tc.expectedIssues {
				found := false
				for _, failure := range result.Error {
					if failure.Text == expectedIssue {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected to find issue: %s", expectedIssue)
			}
		})
	}
}

func TestIngressAnalyzerLabelSelector(t *testing.T) {
	clientSet := fake.NewSimpleClientset(
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress-with-label",
				Namespace: "default",
				Labels: map[string]string{
					"app": "test",
				},
			},
			Spec: networkingv1.IngressSpec{
				// Missing ingress class to trigger a failure
			},
		},
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress-without-label",
				Namespace: "default",
			},
			Spec: networkingv1.IngressSpec{
				// Missing ingress class to trigger a failure
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

	analyzer := IngressAnalyzer{}
	results, err := analyzer.Analyze(config)
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, "default/ingress-with-label", results[0].Name)
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func pathTypePtr(p networkingv1.PathType) *networkingv1.PathType {
	return &p
}
