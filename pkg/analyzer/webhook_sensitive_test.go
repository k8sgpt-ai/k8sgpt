/*
Copyright 2026 The K8sGPT Authors.
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
	"fmt"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	"github.com/stretchr/testify/require"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func webhookMaskingObjects(mutating bool) []runtime.Object {
	objects := []runtime.Object{
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "emptysvc",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"app": "nopods"},
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "podsvc",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"app": "pending"},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pendingpod",
				Namespace: "default",
				Labels:    map[string]string{"app": "pending"},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
			},
		},
	}
	webhooks := []struct {
		name        string
		serviceName string
	}{
		{name: "svcmisshook", serviceName: "missingsvc"},
		{name: "nopodshook", serviceName: "emptysvc"},
		{name: "podhook", serviceName: "podsvc"},
	}
	if mutating {
		mutatingWebhooks := make([]admissionregistrationv1.MutatingWebhook, 0, len(webhooks))
		for _, wh := range webhooks {
			mutatingWebhooks = append(mutatingWebhooks, admissionregistrationv1.MutatingWebhook{
				Name: wh.name,
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Name:      wh.serviceName,
						Namespace: "default",
					},
				},
			})
		}
		objects = append(objects, &admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{Name: "webhookconfig"},
			Webhooks:   mutatingWebhooks,
		})
	} else {
		validatingWebhooks := make([]admissionregistrationv1.ValidatingWebhook, 0, len(webhooks))
		for _, wh := range webhooks {
			validatingWebhooks = append(validatingWebhooks, admissionregistrationv1.ValidatingWebhook{
				Name: wh.name,
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Name:      wh.serviceName,
						Namespace: "default",
					},
				},
			})
		}
		objects = append(objects, &admissionregistrationv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{Name: "webhookconfig"},
			Webhooks:   validatingWebhooks,
		})
	}
	return objects
}

func TestWebhookAnalyzerSensitiveMasking(t *testing.T) {
	for _, tt := range []struct {
		name     string
		mutating bool
		kind     string
	}{
		{name: "mutating webhook", mutating: true, kind: "Mutating"},
		{name: "validating webhook", mutating: false, kind: "Validating"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			config := common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(webhookMaskingObjects(tt.mutating)...),
				},
				Context:   context.Background(),
				Namespace: "default",
			}
			var results []common.Result
			var err error
			if tt.mutating {
				mwAnalyzer := MutatingWebhookAnalyzer{}
				results, err = mwAnalyzer.Analyze(config)
			} else {
				vwAnalyzer := ValidatingWebhookAnalyzer{}
				results, err = vwAnalyzer.Analyze(config)
			}
			require.NoError(t, err)
			require.Len(t, results, 3)

			wantTexts := []string{
				fmt.Sprintf("Service missingsvc not found as mapped to by %s Webhook svcmisshook", tt.kind),
				fmt.Sprintf("No active pods found within service emptysvc as mapped to by %s Webhook nopodshook", tt.kind),
				fmt.Sprintf("%s Webhook (podhook) is pointing to an inactive receiver pod (pendingpod)", tt.kind),
			}
			leaked := []string{"missingsvc", "svcmisshook", "emptysvc", "nopodshook", "podhook", "pendingpod"}
			for _, wantText := range wantTexts {
				var failure *common.Failure
				for i := range results {
					for j := range results[i].Error {
						if results[i].Error[j].Text == wantText {
							failure = &results[i].Error[j]
						}
					}
				}
				require.NotNil(t, failure, "Expected message %q not found in analysis results", wantText)

				for _, s := range failure.Sensitive {
					require.Contains(t, failure.Text, s.Unmasked, "Sensitive.Unmasked %q not found in failure text %q; masking is ineffective", s.Unmasked, failure.Text)
				}
				masked := failure.Text
				for _, s := range failure.Sensitive {
					masked = util.ReplaceIfMatch(masked, s.Unmasked, s.Masked)
				}
				for _, value := range leaked {
					require.NotContains(t, masked, value, "value %q leaked after masking: %q", value, masked)
				}
			}
		})
	}
}
