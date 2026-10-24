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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

func olmObject(gvr schema.GroupVersionResource, kind, namespace, name string, status map[string]any) *unstructured.Unstructured {
	obj := map[string]any{
		"apiVersion": gvr.GroupVersion().String(),
		"kind":       kind,
		"metadata": map[string]any{
			"name":      name,
			"namespace": namespace,
			"labels":    map[string]any{"app": "web"},
		},
	}
	if status != nil {
		obj["status"] = status
	}
	return &unstructured.Unstructured{Object: obj}
}

func TestOLMAnalyzersHonorAnalyzerFilters(t *testing.T) {
	tests := []struct {
		kind     string
		gvr      schema.GroupVersionResource
		analyzer common.IAnalyzer
		status   map[string]any
	}{
		{
			kind:     "CatalogSource",
			gvr:      catSrcGVR,
			analyzer: CatalogSourceAnalyzer{},
			status:   map[string]any{"connectionState": map[string]any{"lastObservedState": "TRANSIENT_FAILURE"}},
		},
		{
			kind:     "ClusterServiceVersion",
			gvr:      csvGVR,
			analyzer: ClusterServiceVersionAnalyzer{},
			status:   map[string]any{"phase": "Failed"},
		},
		{
			kind:     "InstallPlan",
			gvr:      ipGVR,
			analyzer: InstallPlanAnalyzer{},
			status:   map[string]any{"phase": "Failed"},
		},
		{
			kind:     "Subscription",
			gvr:      subGVR,
			analyzer: SubscriptionAnalyzer{},
			status:   map[string]any{"state": "UpgradePending"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			dc := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
				runtime.NewScheme(),
				map[schema.GroupVersionResource]string{tt.gvr: tt.kind + "List"},
				olmObject(tt.gvr, tt.kind, "team-a", "web", tt.status),
				olmObject(tt.gvr, tt.kind, "team-b", "web", tt.status),
			)
			a := common.Analyzer{
				Context:       context.Background(),
				Client:        &kubernetes.Client{DynamicClient: dc},
				Namespace:     "team-a",
				LabelSelector: "app=web",
				ResourceName:  "web",
			}

			res, err := tt.analyzer.Analyze(a)
			if err != nil {
				t.Fatalf("Analyze error: %v", err)
			}
			if len(res) != 1 || res[0].Name != "team-a/web" {
				t.Fatalf("expected only team-a/web, got %#v", res)
			}

			var list k8stesting.ListAction
			for _, action := range dc.Actions() {
				if l, ok := action.(k8stesting.ListAction); ok {
					list = l
				}
			}
			if list == nil {
				t.Fatal("expected a list call")
			}
			if got := list.GetNamespace(); got != "team-a" {
				t.Errorf("list namespace = %q, want %q", got, "team-a")
			}
			restrictions := list.GetListRestrictions()
			if got := restrictions.Labels.String(); got != "app=web" {
				t.Errorf("list label selector = %q, want %q", got, "app=web")
			}
			if got := restrictions.Fields.String(); got != "metadata.name=web" {
				t.Errorf("list field selector = %q, want %q", got, "metadata.name=web")
			}
		})
	}
}

func TestOperatorGroupAnalyzerHonorsNamespace(t *testing.T) {
	dc := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{ogGVR: "OperatorGroupList"},
		olmObject(ogGVR, "OperatorGroup", "team-a", "og-1", nil),
		olmObject(ogGVR, "OperatorGroup", "team-a", "og-2", nil),
		olmObject(ogGVR, "OperatorGroup", "team-b", "og-1", nil),
		olmObject(ogGVR, "OperatorGroup", "team-b", "og-2", nil),
	)
	a := common.Analyzer{
		Context:   context.Background(),
		Client:    &kubernetes.Client{DynamicClient: dc},
		Namespace: "team-a",
	}

	res, err := (OperatorGroupAnalyzer{}).Analyze(a)
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}
	if len(res) != 1 || res[0].Name != "team-a" {
		t.Fatalf("expected only team-a, got %#v", res)
	}
}
