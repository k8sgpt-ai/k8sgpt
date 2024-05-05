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
	validIgClassName := new(string)
	*validIgClassName = "valid-ingress-class"

	var igRule networkingv1.IngressRule

	httpRule := networkingv1.HTTPIngressRuleValue{
		Paths: []networkingv1.HTTPIngressPath{
			{
				Path: "/",
				Backend: networkingv1.IngressBackend{
					Service: &networkingv1.IngressServiceBackend{
						// This service exists.
						Name: "Service1",
					},
				},
			},
			{
				Path: "/test1",
				Backend: networkingv1.IngressBackend{
					Service: &networkingv1.IngressServiceBackend{
						// This service is in the test namespace
						// Hence, it won't be discovered.
						Name: "Service2",
					},
				},
			},
			{
				Path: "/test2",
				Backend: networkingv1.IngressBackend{
					Service: &networkingv1.IngressServiceBackend{
						// This service doesn't exist.
						Name: "Service3",
					},
				},
			},
		},
	}
	igRule.IngressRuleValue.HTTP = &httpRule

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: fake.NewSimpleClientset(
				&networkingv1.Ingress{
					// Doesn't specify an ingress class.
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Ingress1",
						Namespace: "default",
					},
				},
				&networkingv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Ingress2",
						Namespace: "default",
						// Specify an invalid ingress class name using annotations.
						Annotations: map[string]string{
							"kubernetes.io/ingress.class": "invalid-class",
						},
					},
				},
				&networkingv1.Ingress{
					// Namespace filtering.
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Ingress3",
						Namespace: "test",
					},
				},
				&networkingv1.IngressClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: *validIgClassName,
					},
				},
				&networkingv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Ingress4",
						Namespace: "default",
						// Specify valid ingress class name using annotations.
						Annotations: map[string]string{
							"kubernetes.io/ingress.class": *validIgClassName,
						},
					},
				},
				&v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Service1",
						Namespace: "default",
					},
				},
				&v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						// Namespace filtering.
						Name:      "Service2",
						Namespace: "test",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Secret1",
						Namespace: "default",
					},
				},
				&v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Secret2",
						Namespace: "test",
					},
				},
				&networkingv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Ingress5",
						Namespace: "default",
					},

					// Specify valid ingress class name in spec.
					Spec: networkingv1.IngressSpec{
						IngressClassName: validIgClassName,
						Rules: []networkingv1.IngressRule{
							igRule,
						},
						TLS: []networkingv1.IngressTLS{
							{
								// This won't contribute to any failures.
								SecretName: "Secret1",
							},
							{
								// This secret won't be discovered because of namespace filtering.
								SecretName: "Secret2",
							},
							{
								// This secret doesn't exist.
								SecretName: "Secret3",
							},
						},
					},
				},
			),
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	igAnalyzer := IngressAnalyzer{}
	results, err := igAnalyzer.Analyze(config)
	require.NoError(t, err)

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	expectations := []struct {
		name          string
		failuresCount int
	}{
		{
			name:          "default/Ingress1",
			failuresCount: 1,
		},
		{
			name:          "default/Ingress2",
			failuresCount: 1,
		},
		{
			name:          "default/Ingress5",
			failuresCount: 4,
		},
	}

	require.Equal(t, len(expectations), len(results))

	for i, result := range results {
		require.Equal(t, expectations[i].name, result.Name)
		require.Equal(t, expectations[i].failuresCount, len(result.Error))
	}
}
