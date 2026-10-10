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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestAnalyzeRoleBindingsResolvesClusterRoleReferences(t *testing.T) {
	const namespace = "team-a"

	tests := []struct {
		name             string
		clusterRoleRules []rbacv1.PolicyRule
		roleRules        []rbacv1.PolicyRule
		expectedResults  int
	}{
		{
			name: "reports wildcard permissions from cluster role",
			clusterRoleRules: []rbacv1.PolicyRule{{
				Verbs:     []string{"*"},
				Resources: []string{"pods"},
			}},
			expectedResults: 1,
		},
		{
			name: "does not inspect same-named namespaced role",
			clusterRoleRules: []rbacv1.PolicyRule{{
				Verbs:     []string{"get"},
				Resources: []string{"pods"},
			}},
			roleRules: []rbacv1.PolicyRule{{
				Verbs:     []string{"*"},
				Resources: []string{"secrets"},
			}},
			expectedResults: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterRole := &rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{Name: "shared-access"},
				Rules:      tt.clusterRoleRules,
			}
			roleBinding := &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{Name: "team-binding", Namespace: namespace},
				RoleRef: rbacv1.RoleRef{
					Kind: "ClusterRole",
					Name: "shared-access",
				},
			}

			objects := []runtime.Object{clusterRole, roleBinding}
			if tt.roleRules != nil {
				objects = append(objects, &rbacv1.Role{
					ObjectMeta: metav1.ObjectMeta{Name: "shared-access", Namespace: namespace},
					Rules:      tt.roleRules,
				})
			}

			client := fake.NewSimpleClientset(objects...)
			results, err := analyzeRoleBindings(common.Analyzer{
				Client:    &kubernetes.Client{Client: client},
				Context:   context.Background(),
				Namespace: namespace,
			})

			require.NoError(t, err)
			require.Len(t, results, tt.expectedResults)
			if tt.expectedResults == 1 {
				require.Equal(t, "Security/RoleBinding", results[0].Kind)
				require.Equal(t, "team-a/team-binding", results[0].Name)
				require.Contains(t, results[0].Error[0].Text, "ClusterRole shared-access")
			}
		})
	}
}
