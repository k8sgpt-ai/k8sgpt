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
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/stretchr/testify/require"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestValidatingWebhookAnalyzer(t *testing.T) {
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: fake.NewSimpleClientset(
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Pod1",
						Namespace: "default",
						Labels: map[string]string{
							"pod": "Pod1",
						},
					},
				},
				&v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service1",
						Namespace: "default",
					},
					Spec: v1.ServiceSpec{
						Selector: map[string]string{
							"pod": "Pod1",
						},
					},
				},
				&v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service2",
						Namespace: "test",
					},
					Spec: v1.ServiceSpec{
						// No such pod exists in the test namespace
						Selector: map[string]string{
							"pod": "Pod2",
						},
					},
				},
				&v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service3",
						Namespace: "test",
					},
					Spec: v1.ServiceSpec{
						// len(service.Spec.Selector) == 0
						Selector: map[string]string{},
					},
				},
				&admissionregistrationv1.ValidatingWebhookConfiguration{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-validating-webhook-config",
						Namespace: "test",
					},
					Webhooks: []admissionregistrationv1.ValidatingWebhook{
						{
							// Failure: Pointing to an inactive receiver pod
							Name: "webhook1",
							ClientConfig: admissionregistrationv1.WebhookClientConfig{
								Service: &admissionregistrationv1.ServiceReference{
									Name:      "test-service1",
									Namespace: "default",
								},
							},
						},
						{
							// Failure: No active pods found in the test namespace
							Name: "webhook2",
							ClientConfig: admissionregistrationv1.WebhookClientConfig{
								Service: &admissionregistrationv1.ServiceReference{
									Name:      "test-service2",
									Namespace: "test",
								},
							},
						},
						{
							Name: "webhook3",
							ClientConfig: admissionregistrationv1.WebhookClientConfig{
								Service: &admissionregistrationv1.ServiceReference{
									Name:      "test-service3",
									Namespace: "test",
								},
							},
						},
						{
							// Failure: Service doesn't exist.
							Name: "webhook4",
							ClientConfig: admissionregistrationv1.WebhookClientConfig{
								Service: &admissionregistrationv1.ServiceReference{
									Name:      "test-service4-doesn't-exist",
									Namespace: "test",
								},
							},
						},
						{
							// Service is nil.
							Name:         "webhook5",
							ClientConfig: admissionregistrationv1.WebhookClientConfig{},
						},
					},
				},
			),
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	vwAnalyzer := ValidatingWebhookAnalyzer{}
	results, err := vwAnalyzer.Analyze(config)
	require.NoError(t, err)

	// The results should contain: webhook1, webhook2, and webhook4
	resultsLen := 3
	require.Equal(t, resultsLen, len(results))
}
