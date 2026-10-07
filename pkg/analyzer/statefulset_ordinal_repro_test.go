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
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestStatefulSetAnalyzerUsesConfiguredStartOrdinal(t *testing.T) {
	tests := []struct {
		name          string
		pods          []*corev1.Pod
		events        []*corev1.Event
		expectedError string
	}{
		{
			name: "reports a non-zero ordinal pod",
			pods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "database-1", Namespace: "default"},
					Status:     corev1.PodStatus{Phase: corev1.PodPending},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "database-2", Namespace: "default"},
					Status:     corev1.PodStatus{Phase: corev1.PodRunning},
				},
			},
			expectedError: "Statefulset pod database-1 in the namespace default is not in running state.",
		},
		{
			name: "uses the non-zero first ordinal for missing pod events",
			pods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "database-2", Namespace: "default"},
					Status:     corev1.PodStatus{Phase: corev1.PodRunning},
				},
			},
			events: []*corev1.Event{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "database-warning", Namespace: "default"},
					InvolvedObject: corev1.ObjectReference{
						Kind:      "StatefulSet",
						Namespace: "default",
						Name:      "database",
					},
					Type:    "Warning",
					Message: "database-1 could not be created",
				},
			},
			expectedError: "database-1 could not be created",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			replicas := int32(2)
			objects := []runtime.Object{
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "database", Namespace: "default"},
					Spec: appsv1.StatefulSetSpec{
						Replicas:    &replicas,
						ServiceName: "database",
						Ordinals:    &appsv1.StatefulSetOrdinals{Start: 1},
					},
					Status: appsv1.StatefulSetStatus{AvailableReplicas: 1},
				},
				&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "database", Namespace: "default"}},
			}
			for _, pod := range tt.pods {
				objects = append(objects, pod)
			}
			for _, event := range tt.events {
				objects = append(objects, event)
			}

			clientset := fake.NewSimpleClientset(objects...)
			results, err := (StatefulSetAnalyzer{}).Analyze(common.Analyzer{
				Client:    &kubernetes.Client{Client: clientset},
				Context:   context.Background(),
				Namespace: "default",
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(results) != 1 || len(results[0].Error) != 1 || results[0].Error[0].Text != tt.expectedError {
				t.Fatalf("expected one error %q, got %#v", tt.expectedError, results)
			}
		})
	}
}
