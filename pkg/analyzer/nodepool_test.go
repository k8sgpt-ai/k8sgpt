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

// newNodePool builds a fake NodePool. Any of limits, resources and conditions
// may be nil to leave that field out.
func newNodePool(name string, limits any, resources any, conditions any) *unstructured.Unstructured {
	obj := map[string]any{
		"apiVersion": "karpenter.sh/v1",
		"kind":       "NodePool",
		"metadata":   map[string]any{"name": name},
	}
	if limits != nil {
		obj["spec"] = map[string]any{"limits": limits}
	}
	status := map[string]any{}
	if resources != nil {
		status["resources"] = resources
	}
	if conditions != nil {
		status["conditions"] = conditions
	}
	if len(status) > 0 {
		obj["status"] = status
	}
	return &unstructured.Unstructured{Object: obj}
}

func readyCond(status, reason, message string) []any {
	c := map[string]any{"type": "Ready", "status": status}
	if reason != "" {
		c["reason"] = reason
	}
	if message != "" {
		c["message"] = message
	}
	return []any{c}
}

func newNodePoolFakeClient(objs ...runtime.Object) *dynamicfake.FakeDynamicClient {
	listKinds := map[schema.GroupVersionResource]string{
		nodePoolGVR: "NodePoolList",
	}
	return dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), listKinds, objs...)
}

func TestNodePoolAnalyzer(t *testing.T) {
	ready := readyCond("True", "", "")

	tests := []struct {
		name         string
		objects      []runtime.Object
		wantResults  int
		wantName     string   // name of the first result, when set
		wantFailures []string // substrings expected in each failure of the first result, in order
	}{
		{
			name: "ready and under limits",
			objects: []runtime.Object{newNodePool("ok",
				map[string]any{"cpu": "1000", "memory": "1000Gi"},
				map[string]any{"cpu": "8", "memory": "32Gi"},
				ready)},
			wantResults: 0,
		},
		{
			name:        "ready with no limits",
			objects:     []runtime.Object{newNodePool("nolimits", nil, map[string]any{"cpu": "8"}, ready)},
			wantResults: 0,
		},
		{
			name: "ready false with reason and message",
			objects: []runtime.Object{newNodePool("default", nil, nil,
				readyCond("False", "NodeClassNotReady", "EC2NodeClass default is not ready"))},
			wantResults:  1,
			wantName:     "default",
			wantFailures: []string{"Ready=False: NodeClassNotReady: EC2NodeClass default is not ready"},
		},
		{
			name:         "ready false without reason or message",
			objects:      []runtime.Object{newNodePool("bare", nil, nil, readyCond("False", "", ""))},
			wantResults:  1,
			wantFailures: []string{"NodePool bare condition Ready=False"},
		},
		{
			name:        "ready unknown is not reported",
			objects:     []runtime.Object{newNodePool("new", nil, nil, readyCond("Unknown", "", ""))},
			wantResults: 0,
		},
		{
			name:        "no status",
			objects:     []runtime.Object{newNodePool("fresh", map[string]any{"cpu": "100"}, nil, nil)},
			wantResults: 0,
		},
		{
			name: "cpu limit reached exactly",
			objects: []runtime.Object{newNodePool("cpu-full",
				map[string]any{"cpu": "100"},
				map[string]any{"cpu": "100"},
				ready)},
			wantResults:  1,
			wantFailures: []string{"has reached its cpu limit (100 used of 100)"},
		},
		{
			name: "usage above limit is reported",
			objects: []runtime.Object{newNodePool("over",
				map[string]any{"memory": "1000Gi"},
				map[string]any{"memory": "1100Gi"},
				ready)},
			wantResults:  1,
			wantFailures: []string{"memory limit"},
		},
		{
			name: "limit compared across units",
			objects: []runtime.Object{newNodePool("units",
				map[string]any{"memory": "1Ti"},
				map[string]any{"memory": "1024Gi"},
				ready)},
			wantResults:  1,
			wantFailures: []string{"memory limit"},
		},
		{
			name: "numeric limit",
			objects: []runtime.Object{newNodePool("numeric",
				map[string]any{"cpu": int64(10)},
				map[string]any{"cpu": "10"},
				ready)},
			wantResults:  1,
			wantFailures: []string{"cpu limit (10 used of 10)"},
		},
		{
			name: "zero limit is skipped",
			objects: []runtime.Object{newNodePool("disabled",
				map[string]any{"cpu": "0"},
				map[string]any{"cpu": "0"},
				ready)},
			wantResults: 0,
		},
		{
			name: "limited resource missing from usage",
			objects: []runtime.Object{newNodePool("gpu",
				map[string]any{"nvidia.com/gpu": "8"},
				map[string]any{"cpu": "8"},
				ready)},
			wantResults: 0,
		},
		{
			name: "unparseable quantity is skipped",
			objects: []runtime.Object{newNodePool("garbage",
				map[string]any{"cpu": "lots"},
				map[string]any{"cpu": "10"},
				ready)},
			wantResults: 0,
		},
		{
			name: "limits is not a map",
			objects: []runtime.Object{newNodePool("bad-limits",
				"oops",
				map[string]any{"cpu": "10"},
				ready)},
			wantResults: 0,
		},
		{
			name: "conditions is not a list",
			objects: []runtime.Object{newNodePool("bad-conditions", nil, nil,
				"oops")},
			wantResults: 0,
		},
		{
			name: "multiple limits reached are reported in sorted order",
			objects: []runtime.Object{newNodePool("full",
				map[string]any{"memory": "64Gi", "cpu": "16"},
				map[string]any{"memory": "64Gi", "cpu": "16"},
				ready)},
			wantResults:  1,
			wantFailures: []string{"cpu limit", "memory limit"},
		},
		{
			name: "not ready and limit reached on the same pool",
			objects: []runtime.Object{newNodePool("both",
				map[string]any{"cpu": "16"},
				map[string]any{"cpu": "16"},
				readyCond("False", "NodeClassNotReady", ""))},
			wantResults:  1,
			wantFailures: []string{"Ready=False", "cpu limit"},
		},
		{
			name:        "no nodepools",
			objects:     nil,
			wantResults: 0,
		},
		{
			name: "only the failing nodepool is reported",
			objects: []runtime.Object{
				newNodePool("healthy", map[string]any{"cpu": "1000"}, map[string]any{"cpu": "8"}, ready),
				newNodePool("broken", nil, nil, readyCond("False", "NodeClassNotReady", "")),
			},
			wantResults:  1,
			wantName:     "broken",
			wantFailures: []string{"Ready=False"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := common.Analyzer{
				Context: context.TODO(),
				Client:  &kubernetes.Client{DynamicClient: newNodePoolFakeClient(tt.objects...)},
			}

			results, err := NodePoolAnalyzer{}.Analyze(a)
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
			if got.Kind != "NodePool" {
				t.Errorf("expected kind NodePool, got %q", got.Kind)
			}
			if tt.wantName != "" && got.Name != tt.wantName {
				t.Errorf("expected name %q, got %q", tt.wantName, got.Name)
			}
			if len(got.Error) != len(tt.wantFailures) {
				t.Fatalf("expected %d failures, got %d: %#v", len(tt.wantFailures), len(got.Error), got.Error)
			}
			for i, want := range tt.wantFailures {
				if !strings.Contains(got.Error[i].Text, want) {
					t.Errorf("failure %d: expected text to contain %q, got %q", i, want, got.Error[i].Text)
				}
				if len(got.Error[i].Sensitive) == 0 || got.Error[i].Sensitive[0].Unmasked != got.Name {
					t.Errorf("failure %d: expected nodepool name to be marked sensitive, got %#v", i, got.Error[i].Sensitive)
				}
			}
		})
	}
}

func TestNodePoolAnalyzer_CRDNotInstalled(t *testing.T) {
	client := newNodePoolFakeClient()
	client.PrependReactor("list", "nodepools", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewNotFound(nodePoolGVR.GroupResource(), "")
	})

	a := common.Analyzer{
		Context: context.TODO(),
		Client:  &kubernetes.Client{DynamicClient: client},
	}

	_, err := NodePoolAnalyzer{}.Analyze(a)
	if err == nil {
		t.Fatal("expected an error when the NodePool CRD is not installed")
	}
	if !strings.Contains(err.Error(), "karpenter.sh/v1 NodePool: CRD not installed") {
		t.Errorf("expected a CRD-not-installed error naming the API, got %v", err)
	}
	if !apierrors.IsNotFound(err) {
		t.Errorf("expected the wrapped error to still satisfy apierrors.IsNotFound, got %v", err)
	}
}

func TestNodePoolAnalyzer_NilDynamicClient(t *testing.T) {
	a := common.Analyzer{
		Context: context.TODO(),
		Client:  &kubernetes.Client{},
	}

	if _, err := (NodePoolAnalyzer{}).Analyze(a); err == nil {
		t.Fatal("expected an error for a nil dynamic client")
	}
}
