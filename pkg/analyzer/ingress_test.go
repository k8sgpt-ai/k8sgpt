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
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestIngressAnalyzer(t *testing.T) {
	tests := []struct {
		name         string
		config       common.Analyzer
		expectations []struct {
			name          string
			failuresCount int
		}
	}{
		{
			name: "Missing ingress class",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&networkingv1.Ingress{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "no-class",
								Namespace: "default",
							},
							Spec: networkingv1.IngressSpec{
								// No ingress class specified
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
					name:          "default/no-class",
					failuresCount: 1, // One failure for missing ingress class
				},
			},
		},
		{
			name: "Non-existent ingress class",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&networkingv1.Ingress{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "bad-class",
								Namespace: "default",
							},
							Spec: networkingv1.IngressSpec{
								IngressClassName: strPtr("non-existent"),
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
					name:          "default/bad-class",
					failuresCount: 1, // One failure for non-existent ingress class
				},
			},
		},
		{
			name: "Non-existent backend service",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&networkingv1.Ingress{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "bad-backend",
								Namespace: "default",
								Annotations: map[string]string{
									"kubernetes.io/ingress.class": "nginx",
								},
							},
							Spec: networkingv1.IngressSpec{
								Rules: []networkingv1.IngressRule{
									{
										Host: "example.com",
										IngressRuleValue: networkingv1.IngressRuleValue{
											HTTP: &networkingv1.HTTPIngressRuleValue{
												Paths: []networkingv1.HTTPIngressPath{
													{
														Path:     "/",
														PathType: pathTypePtr(networkingv1.PathTypePrefix),
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
					name:          "default/bad-backend",
					failuresCount: 2, // Two failures: non-existent ingress class and non-existent service
				},
			},
		},
		{
			name: "Non-existent TLS secret",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&networkingv1.Ingress{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "bad-tls",
								Namespace: "default",
								Annotations: map[string]string{
									"kubernetes.io/ingress.class": "nginx",
								},
							},
							Spec: networkingv1.IngressSpec{
								TLS: []networkingv1.IngressTLS{
									{
										Hosts:      []string{"example.com"},
										SecretName: "non-existent-secret",
									},
								},
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
					name:          "default/bad-tls",
					failuresCount: 2, // Two failures: non-existent ingress class and non-existent TLS secret
				},
			},
		},
		{
			name: "Valid ingress with all components",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&networkingv1.Ingress{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "valid-ingress",
								Namespace: "default",
							},
							Spec: networkingv1.IngressSpec{
								IngressClassName: strPtr("nginx"),
							},
						},
						&networkingv1.IngressClass{
							ObjectMeta: metav1.ObjectMeta{
								Name: "nginx",
							},
						},
						&v1.Service{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "backend-service",
								Namespace: "default",
							},
						},
						&v1.Secret{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "tls-secret",
								Namespace: "default",
							},
							Type: v1.SecretTypeTLS,
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
				// No expectations for valid ingress
			},
		},
		{
			name: "Multiple issues",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&networkingv1.Ingress{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "multiple-issues",
								Namespace: "default",
							},
							Spec: networkingv1.IngressSpec{
								IngressClassName: strPtr("non-existent"),
								Rules: []networkingv1.IngressRule{
									{
										Host: "example.com",
										IngressRuleValue: networkingv1.IngressRuleValue{
											HTTP: &networkingv1.HTTPIngressRuleValue{
												Paths: []networkingv1.HTTPIngressPath{
													{
														Path:     "/",
														PathType: pathTypePtr(networkingv1.PathTypePrefix),
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
								TLS: []networkingv1.IngressTLS{
									{
										Hosts:      []string{"example.com"},
										SecretName: "non-existent-secret",
									},
								},
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
					failuresCount: 3, // Three failures: ingress class, service, and TLS secret
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := IngressAnalyzer{}
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
