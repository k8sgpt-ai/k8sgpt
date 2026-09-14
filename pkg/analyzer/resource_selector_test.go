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
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

// unhealthyDeployment is a deployment the DeploymentAnalyzer will report on
// (available replicas below desired), so each list result turns into a finding.
func unhealthyDeployment(name string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "demo"},
		Spec: appsv1.DeploymentSpec{
			Replicas: func() *int32 { i := int32(3); return &i }(),
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{Containers: []v1.Container{{Name: "c", Image: "nginx"}}},
			},
		},
		Status: appsv1.DeploymentStatus{Replicas: 2, AvailableReplicas: 1},
	}
}

// The fake clientset does NOT implement field selectors: it returns every object
// regardless of ListOptions.FieldSelector. This reactor stands in for the API
// server so the test exercises the real contract - the analyzer asks for one
// object by name, and only that object comes back. capturedFieldSelector also
// lets the test assert the predicate really was pushed down to the list call
// rather than applied after a wide list.
func fieldSelectorAwareClient(objs ...*appsv1.Deployment) (*fake.Clientset, *string) {
	cs := fake.NewSimpleClientset()
	tracker := make([]*appsv1.Deployment, 0, len(objs))
	tracker = append(tracker, objs...)
	captured := new(string)

	cs.PrependReactor("list", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		listAction, ok := action.(k8stesting.ListAction)
		if !ok {
			return false, nil, nil
		}
		fieldSel := listAction.GetListRestrictions().Fields
		*captured = fieldSel.String()

		out := &appsv1.DeploymentList{}
		for _, d := range tracker {
			// mimic the API server: honour a metadata.name field selector
			if fieldSel != nil && !fieldSel.Empty() {
				if v, ok := fieldSel.RequiresExactMatch("metadata.name"); ok && v != d.Name {
					continue
				}
			}
			out.Items = append(out.Items, *d)
		}
		return true, out, nil
	})
	return cs, captured
}

func runDeploymentAnalyzer(t *testing.T, cs *fake.Clientset, resourceName string) []common.Result {
	t.Helper()
	config := common.Analyzer{
		Client:       &kubernetes.Client{Client: cs},
		Context:      context.Background(),
		Namespace:    "demo",
		ResourceName: resourceName,
	}
	results, err := DeploymentAnalyzer{}.Analyze(config)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}
	return results
}

// --resource narrows the analysis to the named object only.
func TestResourceSelector_SelectedKindOnly(t *testing.T) {
	cs, captured := fieldSelectorAwareClient(
		unhealthyDeployment("alpha-api"),
		unhealthyDeployment("beta-worker"),
	)
	results := runDeploymentAnalyzer(t, cs, "alpha-api")

	if want := "metadata.name=alpha-api"; *captured != want {
		t.Errorf("field selector sent to the API server = %q, want %q", *captured, want)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1 (only the requested deployment)", len(results))
	}
	if results[0].Name != "demo/alpha-api" {
		t.Errorf("analyzed %q, want %q", results[0].Name, "demo/alpha-api")
	}
}

// A name that matches nothing yields no results and no error.
func TestResourceSelector_NoMatch(t *testing.T) {
	cs, captured := fieldSelectorAwareClient(
		unhealthyDeployment("alpha-api"),
		unhealthyDeployment("beta-worker"),
	)
	results := runDeploymentAnalyzer(t, cs, "does-not-exist")

	if want := "metadata.name=does-not-exist"; *captured != want {
		t.Errorf("field selector sent to the API server = %q, want %q", *captured, want)
	}
	if len(results) != 0 {
		t.Fatalf("got %d results, want 0 for a name that matches nothing", len(results))
	}
}

// Without the flag, behaviour is exactly as before: no field selector, all
// objects analyzed.
func TestResourceSelector_AbsentPreservesBehaviour(t *testing.T) {
	cs, captured := fieldSelectorAwareClient(
		unhealthyDeployment("alpha-api"),
		unhealthyDeployment("beta-worker"),
	)
	results := runDeploymentAnalyzer(t, cs, "")

	if *captured != "" {
		t.Errorf("field selector = %q, want empty when --resource is not set", *captured)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2 (all deployments analyzed)", len(results))
	}
}
