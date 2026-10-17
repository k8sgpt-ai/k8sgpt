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
	"strings"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

func newNodeClaim(name string, status any) *unstructured.Unstructured {
	obj := map[string]any{
		"apiVersion": "karpenter.sh/v1",
		"kind":       "NodeClaim",
		"metadata":   map[string]any{"name": name},
	}
	if status != nil {
		obj["status"] = status
	}
	return &unstructured.Unstructured{Object: obj}
}

func cond(condType, status, reason, message string) map[string]any {
	c := map[string]any{"type": condType, "status": status}
	if reason != "" {
		c["reason"] = reason
	}
	if message != "" {
		c["message"] = message
	}
	return c
}

func newNodeClaimFakeClient(objs ...runtime.Object) *dynamicfake.FakeDynamicClient {
	listKinds := map[schema.GroupVersionResource]string{
		nodeClaimGVR: "NodeClaimList",
	}
	return dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds, objs...)
}

func TestNodeClaimAnalyzer(t *testing.T) {
	healthy := []any{
		cond("Launched", "True", "", ""),
		cond("Registered", "True", "", ""),
		cond("Initialized", "True", "", ""),
	}

	tests := []struct {
		name         string
		objects      []runtime.Object
		wantResults  int
		wantFailures int      // failures in the first result, when wantResults > 0
		wantName     string   // name of the first result, when set
		wantContains []string // substrings expected in the first failure text
	}{
		{
			name:        "all lifecycle conditions true",
			objects:     []runtime.Object{newNodeClaim("ok", map[string]any{"conditions": healthy})},
			wantResults: 0,
		},
		{
			name: "launched false with reason and message",
			objects: []runtime.Object{newNodeClaim("default-abc12", map[string]any{"conditions": []any{
				cond("Launched", "False", "InsufficientCapacity", "no capacity for m5.large"),
			}})},
			wantResults:  1,
			wantFailures: 1,
			wantName:     "default-abc12",
			wantContains: []string{"Launched=False", "InsufficientCapacity", "no capacity for m5.large"},
		},
		{
			name: "launched false without message is still reported",
			objects: []runtime.Object{newNodeClaim("no-msg", map[string]any{"conditions": []any{
				cond("Launched", "False", "InsufficientCapacity", ""),
			}})},
			wantResults:  1,
			wantFailures: 1,
			wantContains: []string{"Launched=False: InsufficientCapacity"},
		},
		{
			name: "launched false without reason or message is still reported",
			objects: []runtime.Object{newNodeClaim("bare", map[string]any{"conditions": []any{
				cond("Launched", "False", "", ""),
			}})},
			wantResults:  1,
			wantFailures: 1,
			wantContains: []string{"NodeClaim bare condition Launched=False"},
		},
		{
			name: "registered false after launch",
			objects: []runtime.Object{newNodeClaim("reg", map[string]any{"conditions": []any{
				cond("Launched", "True", "", ""),
				cond("Registered", "False", "NodeNotFound", "node did not register"),
			}})},
			wantResults:  1,
			wantFailures: 1,
			wantContains: []string{"Registered=False"},
		},
		{
			name: "multiple false conditions on one nodeclaim give one result",
			objects: []runtime.Object{newNodeClaim("multi", map[string]any{"conditions": []any{
				cond("Launched", "False", "A", "a"),
				cond("Registered", "False", "B", "b"),
			}})},
			wantResults:  1,
			wantFailures: 2,
		},
		{
			name: "unknown status is not reported",
			objects: []runtime.Object{newNodeClaim("pending", map[string]any{"conditions": []any{
				cond("Launched", "Unknown", "", ""),
			}})},
			wantResults: 0,
		},
		{
			name: "untracked condition false is ignored",
			objects: []runtime.Object{newNodeClaim("drift", map[string]any{"conditions": []any{
				cond("Drifted", "False", "", ""),
			}})},
			wantResults: 0,
		},
		{
			name:        "no status yet",
			objects:     []runtime.Object{newNodeClaim("new", nil)},
			wantResults: 0,
		},
		{
			name:        "conditions is not a list",
			objects:     []runtime.Object{newNodeClaim("bad", map[string]any{"conditions": "oops"})},
			wantResults: 0,
		},
		{
			name: "condition element is not a map",
			objects: []runtime.Object{newNodeClaim("bad-elem", map[string]any{"conditions": []any{
				"oops",
				cond("Launched", "False", "InsufficientCapacity", ""),
			}})},
			wantResults:  1,
			wantFailures: 1,
		},
		{
			name:        "no nodeclaims",
			objects:     nil,
			wantResults: 0,
		},
		{
			name: "only the failing nodeclaim is reported",
			objects: []runtime.Object{
				newNodeClaim("healthy", map[string]any{"conditions": healthy}),
				newNodeClaim("broken", map[string]any{"conditions": []any{
					cond("Initialized", "False", "NotReady", "node not ready"),
				}}),
			},
			wantResults:  1,
			wantFailures: 1,
			wantName:     "broken",
			wantContains: []string{"Initialized=False"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := common.Analyzer{
				Context: context.TODO(),
				Client:  &kubernetes.Client{DynamicClient: newNodeClaimFakeClient(tt.objects...)},
			}

			results, err := NodeClaimAnalyzer{}.Analyze(a)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(results) != tt.wantResults {
				t.Fatalf("expected %d results, got %d: %#v", tt.wantResults, len(results), results)
			}
			if tt.wantResults == 0 {
				return
			}

			got := results[0]
			if got.Kind != "NodeClaim" {
				t.Errorf("expected kind NodeClaim, got %q", got.Kind)
			}
			if tt.wantName != "" && got.Name != tt.wantName {
				t.Errorf("expected name %q, got %q", tt.wantName, got.Name)
			}
			if len(got.Error) != tt.wantFailures {
				t.Fatalf("expected %d failures, got %d: %#v", tt.wantFailures, len(got.Error), got.Error)
			}
			for _, s := range tt.wantContains {
				if !strings.Contains(got.Error[0].Text, s) {
					t.Errorf("expected failure text to contain %q, got %q", s, got.Error[0].Text)
				}
			}
			if len(got.Error[0].Sensitive) == 0 || got.Error[0].Sensitive[0].Unmasked != got.Name {
				t.Errorf("expected nodeclaim name to be marked sensitive, got %#v", got.Error[0].Sensitive)
			}
		})
	}
}

func TestNodeClaimAnalyzer_CRDNotInstalled(t *testing.T) {
	client := newNodeClaimFakeClient()
	client.PrependReactor("list", "nodeclaims", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewNotFound(nodeClaimGVR.GroupResource(), "")
	})

	a := common.Analyzer{
		Context: context.TODO(),
		Client:  &kubernetes.Client{DynamicClient: client},
	}

	_, err := NodeClaimAnalyzer{}.Analyze(a)
	if err == nil {
		t.Fatal("expected an error when the NodeClaim CRD is not installed")
	}
	if !strings.Contains(err.Error(), "is Karpenter installed?") {
		t.Errorf("expected a helpful CRD-not-found error, got %v", err)
	}
	if !apierrors.IsNotFound(err) {
		t.Errorf("expected the wrapped error to still satisfy apierrors.IsNotFound, got %v", err)
	}
}

func TestNodeClaimAnalyzer_NilDynamicClient(t *testing.T) {
	a := common.Analyzer{
		Context: context.TODO(),
		Client:  &kubernetes.Client{},
	}

	if _, err := (NodeClaimAnalyzer{}).Analyze(a); err == nil {
		t.Fatal("expected an error for a nil dynamic client")
	}
}
