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
	"fmt"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestClusterExtensionAnalyzer(t *testing.T) {
	gvr := schema.GroupVersionResource{
		Group:    "olm.operatorframework.io",
		Version:  "v1",
		Resource: "clusterextensions",
	}

	scheme := runtime.NewScheme()

	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{
			gvr: "ClusterExtensionList",
		},
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "olm.operatorframework.io/v1",
				"kind":       "ClusterExtension",
				"metadata": map[string]interface{}{
					"name": "Valid SelfCertified ClusterExtension",
				},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"sourceType": "Catalog",
						"catalog": map[string]interface{}{
							"upgradeConstraintPolicy": "SelfCertified",
						},
					},
				},
				"status": map[string]interface{}{
					"conditions": []interface{}{
						map[string]interface{}{
							"type":   "Installed",
							"status": "True",
							"reason": "Succeeded",
						},
					},
				},
			},
		},
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "olm.operatorframework.io/v1",
				"kind":       "ClusterExtension",
				"metadata": map[string]interface{}{
					"name": "Valid CatalogProvided ClusterExtension",
				},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"sourceType": "Catalog",
						"catalog": map[string]interface{}{
							"upgradeConstraintPolicy": "CatalogProvided",
						},
					},
				},
				"status": map[string]interface{}{
					"conditions": []interface{}{
						map[string]interface{}{
							"type":   "Installed",
							"status": "True",
							"reason": "Succeeded",
						},
					},
				},
			},
		},
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "olm.operatorframework.io/v1",
				"kind":       "ClusterExtension",
				"metadata": map[string]interface{}{
					"name": "Invalid UpgradeConstraintPolicy",
				},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"sourceType": "Catalog",
						"catalog": map[string]interface{}{
							"upgradeConstraintPolicy": "InvalidPolicy",
						},
					},
				},
				"status": map[string]interface{}{
					"conditions": []interface{}{
						map[string]interface{}{
							"type":   "Progressing",
							"status": "True",
							"reason": "Retrying",
						},
						map[string]interface{}{
							"type":   "Installed",
							"status": "False",
							"reason": "Failed",
						},
					},
				},
			},
		},
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "olm.operatorframework.io/v1",
				"kind":       "ClusterExtension",
				"metadata": map[string]interface{}{
					"name": "Invalid SourceType",
				},
				"spec": map[string]interface{}{
					"source": map[string]interface{}{
						"sourceType": "Git",
						"catalog": map[string]interface{}{
							"upgradeConstraintPolicy": "CatalogProvided",
						},
					},
				},
				"status": map[string]interface{}{
					"conditions": []interface{}{
						map[string]interface{}{
							"type":   "Progressing",
							"status": "True",
							"reason": "Retrying",
						},
						map[string]interface{}{
							"type":   "Installed",
							"status": "False",
							"reason": "Failed",
						},
					},
				},
			},
		},
	)
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client:        fake.NewSimpleClientset(),
			DynamicClient: dynamicClient,
		},
		Context:   context.Background(),
		Namespace: "test",
	}

	ceAnalyzer := ClusterExtensionAnalyzer{}
	results, err := ceAnalyzer.Analyze(config)
	for _, res := range results {
		fmt.Printf("Result: %s | Failures: %d\n", res.Name, len(res.Error))
		for _, err := range res.Error {
			fmt.Printf("  - %s\n", err)
		}
	}
	require.NoError(t, err)
	require.Equal(t, 2, len(results))
}
