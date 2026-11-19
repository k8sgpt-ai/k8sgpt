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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

// The dynamic fake client does not apply field selectors, and an empty
// ListOptions must be distinguishable from a real filter. This reactor stands
// in for the API server: it returns only the objects the list call asked for.
func selectorAwareDynamic(gvr schema.GroupVersionResource, kind string, objs []*unstructured.Unstructured) *dynamicfake.FakeDynamicClient {
	dc := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{gvr: kind + "List"},
	)
	dc.PrependReactor("list", gvr.Resource, func(action k8stesting.Action) (bool, runtime.Object, error) {
		listAction, ok := action.(k8stesting.ListAction)
		if !ok {
			return false, nil, nil
		}
		restrictions := listAction.GetListRestrictions()
		out := &unstructured.UnstructuredList{}
		for _, obj := range objs {
			if !matchesLabelSelector(restrictions, obj) || !matchesFieldSelector(restrictions, obj) {
				continue
			}
			out.Items = append(out.Items, *obj)
		}
		return true, out, nil
	})
	return dc
}

func matchesLabelSelector(restrictions k8stesting.ListRestrictions, obj *unstructured.Unstructured) bool {
	if restrictions.Labels == nil || restrictions.Labels.Empty() {
		return true
	}
	return restrictions.Labels.Matches(labels.Set(obj.GetLabels()))
}

func matchesFieldSelector(restrictions k8stesting.ListRestrictions, obj *unstructured.Unstructured) bool {
	if restrictions.Fields == nil || restrictions.Fields.Empty() {
		return true
	}
	name, exact := restrictions.Fields.RequiresExactMatch("metadata.name")
	return !exact || name == obj.GetName()
}

func failingClusterExtension(name, app string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "olm.operatorframework.io/v1",
		"kind":       "ClusterExtension",
		"metadata": map[string]interface{}{
			"name":   name,
			"labels": map[string]interface{}{"app": app},
		},
		"spec": map[string]interface{}{
			"source": map[string]interface{}{
				"sourceType": "Git",
				"catalog": map[string]interface{}{
					"upgradeConstraintPolicy": "CatalogProvided",
				},
			},
		},
	}}
}

func failingClusterCatalog(name, app string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "olm.operatorframework.io/v1",
		"kind":       "ClusterCatalog",
		"metadata": map[string]interface{}{
			"name":   name,
			"labels": map[string]interface{}{"app": app},
		},
		// spec.source has no image, so validation reports the catalog.
		"spec": map[string]interface{}{
			"source": map[string]interface{}{"type": "Unknown"},
		},
	}}
}

func TestClusterExtensionAndClusterCatalogHonorFilters(t *testing.T) {
	extensionGVR := schema.GroupVersionResource{Group: "olm.operatorframework.io", Version: "v1", Resource: "clusterextensions"}
	catalogGVR := schema.GroupVersionResource{Group: "olm.operatorframework.io", Version: "v1", Resource: "clustercatalogs"}

	analyzers := []struct {
		kind     string
		gvr      schema.GroupVersionResource
		analyzer common.IAnalyzer
		objects  []*unstructured.Unstructured
	}{
		{
			kind:     "ClusterExtension",
			gvr:      extensionGVR,
			analyzer: ClusterExtensionAnalyzer{},
			objects: []*unstructured.Unstructured{
				failingClusterExtension("web", "web"),
				failingClusterExtension("other", "other"),
			},
		},
		{
			kind:     "ClusterCatalog",
			gvr:      catalogGVR,
			analyzer: ClusterCatalogAnalyzer{},
			objects: []*unstructured.Unstructured{
				failingClusterCatalog("web", "web"),
				failingClusterCatalog("other", "other"),
			},
		},
	}

	cases := []struct {
		name          string
		labelSelector string
		resourceName  string
		wantNames     []string
		wantLabels    string
		wantFields    string
	}{
		{
			name:         "resource name",
			resourceName: "web",
			wantNames:    []string{"web"},
			wantFields:   "metadata.name=web",
		},
		{
			name:          "label selector",
			labelSelector: "app=web",
			wantNames:     []string{"web"},
			wantLabels:    "app=web",
		},
		{
			name:          "name and label",
			labelSelector: "app=web",
			resourceName:  "web",
			wantNames:     []string{"web"},
			wantLabels:    "app=web",
			wantFields:    "metadata.name=web",
		},
		{
			name:          "no match",
			labelSelector: "app=missing",
			wantNames:     nil,
			wantLabels:    "app=missing",
		},
		{
			name:      "unfiltered",
			wantNames: []string{"web", "other"},
		},
	}

	for _, analyzer := range analyzers {
		t.Run(analyzer.kind, func(t *testing.T) {
			for _, tt := range cases {
				t.Run(tt.name, func(t *testing.T) {
					dc := selectorAwareDynamic(analyzer.gvr, analyzer.kind, analyzer.objects)
					res, err := analyzer.analyzer.Analyze(common.Analyzer{
						Context:       context.Background(),
						Client:        &kubernetes.Client{DynamicClient: dc},
						LabelSelector: tt.labelSelector,
						ResourceName:  tt.resourceName,
					})
					if err != nil {
						t.Fatalf("Analyze error: %v", err)
					}

					got := map[string]bool{}
					for _, r := range res {
						if r.Kind != analyzer.kind {
							t.Errorf("result kind = %q, want %q", r.Kind, analyzer.kind)
						}
						got[r.Name] = true
					}
					if len(got) != len(tt.wantNames) {
						t.Errorf("results = %v, want %v", got, tt.wantNames)
					}
					for _, name := range tt.wantNames {
						if !got[name] {
							t.Errorf("missing result %q, got %v", name, got)
						}
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
					restrictions := list.GetListRestrictions()
					if got := restrictions.Labels.String(); got != tt.wantLabels {
						t.Errorf("list label selector = %q, want %q", got, tt.wantLabels)
					}
					if got := restrictions.Fields.String(); got != tt.wantFields {
						t.Errorf("list field selector = %q, want %q", got, tt.wantFields)
					}
				})
			}
		})
	}
}
