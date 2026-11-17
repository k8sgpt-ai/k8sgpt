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

package keda

import (
	"errors"
	"testing"

	kedav1 "github.com/kedacore/keda/v2/apis/keda/v1alpha1"
	"github.com/kedacore/keda/v2/pkg/generated/clientset/versioned/fake"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestDeleteKedaResourcesListErrors(t *testing.T) {
	for _, resource := range []string{"scaledobjects", "scaledjobs", "triggerauthentications", "clustertriggerauthentications"} {
		t.Run(resource, func(t *testing.T) {
			client := fake.NewSimpleClientset(&kedav1.ScaledObject{ObjectMeta: metav1.ObjectMeta{Name: "keep", Namespace: "test"}})
			want := errors.New("listing denied")
			client.PrependReactor("list", resource, func(ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, want
			})
			var err error
			require.NotPanics(t, func() { err = deleteKedaResources(client.KedaV1alpha1()) })
			require.ErrorIs(t, err, want)
			require.Contains(t, err.Error(), resource)
			for _, action := range client.Actions() {
				require.Equal(t, "list", action.GetVerb(), "listing failures must prevent deletions")
			}
		})
	}
}

func TestDeleteKedaResourcesSuccess(t *testing.T) {
	client := fake.NewSimpleClientset(
		&kedav1.ScaledObject{ObjectMeta: metav1.ObjectMeta{Name: "object", Namespace: "test"}},
		&kedav1.ScaledJob{ObjectMeta: metav1.ObjectMeta{Name: "job", Namespace: "test"}},
		&kedav1.TriggerAuthentication{ObjectMeta: metav1.ObjectMeta{Name: "auth", Namespace: "test"}},
		&kedav1.ClusterTriggerAuthentication{ObjectMeta: metav1.ObjectMeta{Name: "cluster-auth"}},
	)
	require.NoError(t, deleteKedaResources(client.KedaV1alpha1()))
	actions := client.Actions()
	require.Len(t, actions, 8)
	for i, action := range actions {
		if i < 4 {
			require.Equal(t, "list", action.GetVerb())
		} else {
			require.Equal(t, "delete", action.GetVerb())
		}
	}
}
