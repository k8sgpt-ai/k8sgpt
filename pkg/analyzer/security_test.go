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
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestSecurityAnalyzer(t *testing.T) {
	tests := []struct {
		name            string
		namespace       string
		serviceAccounts []v1.ServiceAccount
		pods            []v1.Pod
		roles           []rbacv1.Role
		roleBindings    []rbacv1.RoleBinding
		expectedErrors  int
		expectedKinds   []string
	}{
		{
			name:      "default service account usage",
			namespace: "default",
			serviceAccounts: []v1.ServiceAccount{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "default",
						Namespace: "default",
					},
				},
			},
			pods: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-pod",
						Namespace: "default",
					},
					Spec: v1.PodSpec{
						ServiceAccountName: "default",
					},
				},
			},
			expectedErrors: 2,
			expectedKinds:  []string{"Security/ServiceAccount", "Security/Pod"},
		},
		{
			name:      "privileged container",
			namespace: "default",
			pods: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "privileged-pod",
						Namespace: "default",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "privileged-container",
								SecurityContext: &v1.SecurityContext{
									Privileged: boolPtr(true),
								},
							},
						},
					},
				},
			},
			expectedErrors: 1,
			expectedKinds:  []string{"Security/Pod"},
		},
		{
			name:      "wildcard permissions in role",
			namespace: "default",
			roles: []rbacv1.Role{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "wildcard-role",
						Namespace: "default",
					},
					Rules: []rbacv1.PolicyRule{
						{
							Verbs:     []string{"*"},
							Resources: []string{"pods"},
						},
					},
				},
			},
			roleBindings: []rbacv1.RoleBinding{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-binding",
						Namespace: "default",
					},
					RoleRef: rbacv1.RoleRef{
						Kind: "Role",
						Name: "wildcard-role",
					},
				},
			},
			expectedErrors: 1,
			expectedKinds:  []string{"Security/RoleBinding"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()

			// Create test resources
			for _, sa := range tt.serviceAccounts {
				_, err := client.CoreV1().ServiceAccounts(tt.namespace).Create(context.TODO(), &sa, metav1.CreateOptions{})
				assert.NoError(t, err)
			}

			for _, pod := range tt.pods {
				_, err := client.CoreV1().Pods(tt.namespace).Create(context.TODO(), &pod, metav1.CreateOptions{})
				assert.NoError(t, err)
			}

			for _, role := range tt.roles {
				_, err := client.RbacV1().Roles(tt.namespace).Create(context.TODO(), &role, metav1.CreateOptions{})
				assert.NoError(t, err)
			}

			for _, rb := range tt.roleBindings {
				_, err := client.RbacV1().RoleBindings(tt.namespace).Create(context.TODO(), &rb, metav1.CreateOptions{})
				assert.NoError(t, err)
			}

			analyzer := SecurityAnalyzer{}
			results, err := analyzer.Analyze(common.Analyzer{
				Client:    &kubernetes.Client{Client: client},
				Context:   context.TODO(),
				Namespace: tt.namespace,
			})

			assert.NoError(t, err)

			// Debug: Print all results
			t.Logf("Got %d results:", len(results))
			for _, result := range results {
				t.Logf("  Kind: %s, Name: %s", result.Kind, result.Name)
				for _, failure := range result.Error {
					t.Logf("    Failure: %s", failure.Text)
				}
			}

			// Count results by kind
			resultsByKind := make(map[string]int)
			for _, result := range results {
				resultsByKind[result.Kind]++
			}

			// Check that we have the expected number of results for each kind
			for _, expectedKind := range tt.expectedKinds {
				assert.Equal(t, 1, resultsByKind[expectedKind], "Expected 1 result of kind %s", expectedKind)
			}

			// Check total number of results matches expected kinds
			assert.Equal(t, len(tt.expectedKinds), len(results), "Expected %d total results", len(tt.expectedKinds))
		})
	}
}
