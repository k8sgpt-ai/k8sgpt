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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestConfigMapAnalyzer(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		configMaps     []v1.ConfigMap
		pods           []v1.Pod
		expectedErrors int
	}{
		{
			name:      "unused configmap",
			namespace: "default",
			configMaps: []v1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unused-cm",
						Namespace: "default",
					},
					Data: map[string]string{
						"key": "value",
					},
				},
			},
			expectedErrors: 1,
		},
		{
			name:      "empty configmap",
			namespace: "default",
			configMaps: []v1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "empty-cm",
						Namespace: "default",
					},
				},
			},
			expectedErrors: 1,
		},
		{
			name:      "large configmap",
			namespace: "default",
			configMaps: []v1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "large-cm",
						Namespace: "default",
					},
					Data: map[string]string{
						"key": string(make([]byte, 1024*1024+1)), // 1MB + 1 byte
					},
				},
			},
			expectedErrors: 1,
		},
		{
			name:      "used configmap",
			namespace: "default",
			configMaps: []v1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "used-cm",
						Namespace: "default",
					},
					Data: map[string]string{
						"key": "value",
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
						Containers: []v1.Container{
							{
								Name: "test-container",
								EnvFrom: []v1.EnvFromSource{
									{
										ConfigMapRef: &v1.ConfigMapEnvSource{
											LocalObjectReference: v1.LocalObjectReference{
												Name: "used-cm",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()

			// Create test resources
			for _, cm := range tt.configMaps {
				_, err := client.CoreV1().ConfigMaps(tt.namespace).Create(context.TODO(), &cm, metav1.CreateOptions{})
				assert.NoError(t, err)
			}

			for _, pod := range tt.pods {
				_, err := client.CoreV1().Pods(tt.namespace).Create(context.TODO(), &pod, metav1.CreateOptions{})
				assert.NoError(t, err)
			}

			analyzer := ConfigMapAnalyzer{}
			results, err := analyzer.Analyze(common.Analyzer{
				Client:    &kubernetes.Client{Client: client},
				Context:   context.TODO(),
				Namespace: tt.namespace,
			})

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedErrors, len(results))
		})
	}
}

// NEW TEST: TestConfigMapAnalyzer_SidecarPatterns tests known sidecar patterns and skip annotations
func TestConfigMapAnalyzer_SidecarPatterns(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		configMaps     []v1.ConfigMap
		pods           []v1.Pod
		expectedErrors int
	}{
		{
			name:      "grafana dashboard configmap should not be flagged as unused",
			namespace: "monitoring",
			configMaps: []v1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "grafana-dashboard",
						Namespace: "monitoring",
						Labels: map[string]string{
							"grafana_dashboard": "1",
						},
					},
					Data: map[string]string{
						"dashboard. json": `{"title":  "My Dashboard"}`,
					},
				},
			},
			pods:            []v1.Pod{},
			expectedErrors: 0,
		},
		{
			name:      "configmap with skip annotation should be ignored",
			namespace: "default",
			configMaps: []v1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ignored-cm",
						Namespace: "default",
						Annotations: map[string]string{
							"k8sgpt.ai/skip-usage-check": "true",
						},
					},
					Data: map[string]string{
						"key": "value",
					},
				},
			},
			expectedErrors: 0,
		},
		{
			name:      "normal unused configmap should still be flagged",
			namespace: "default",
			configMaps:  []v1.ConfigMap{
				{
					ObjectMeta:  metav1.ObjectMeta{
						Name:      "unused-cm",
						Namespace: "default",
					},
					Data: map[string]string{
						"key":  "value",
					},
				},
			},
			expectedErrors: 1,
		},
		{
			name:      "prometheus rule configmap should not be flagged",
			namespace: "monitoring",
			configMaps: []v1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "prometheus-rules",
						Namespace: "monitoring",
						Labels: map[string]string{
							"prometheus_rule": "1",
						},
					},
					Data: map[string]string{
						"rules.yaml": "groups:  []",
					},
				},
			},
			expectedErrors: 0,
		},
		{
			name:      "custom dynamically-loaded label should work",
			namespace: "default",
			configMaps: []v1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:       "custom-sidecar-cm",
						Namespace: "default",
						Labels: map[string]string{
							"k8sgpt.ai/dynamically-loaded": "true",
						},
					},
					Data:  map[string]string{
						"config":  "value",
					},
				},
			},
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t. Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()

			for _, cm := range tt.configMaps {
				_, err := client.CoreV1().ConfigMaps(tt.namespace).Create(context.TODO(), &cm, metav1.CreateOptions{})
				assert.NoError(t, err)
			}

			for _, pod := range tt.pods {
				_, err := client.CoreV1().Pods(tt.namespace).Create(context.TODO(), &pod, metav1.CreateOptions{})
				assert.NoError(t, err)
			}

			analyzer := ConfigMapAnalyzer{}
			results, err := analyzer.Analyze(common. Analyzer{
				Client:     &kubernetes.Client{Client: client},
				Context:   context.TODO(),
				Namespace: tt.namespace,
			})

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedErrors, len(results), "Expected %d errors but got %d", tt.expectedErrors, len(results))
		})
	}
}

func TestConfigMapAnalyzer_NamespaceScoping(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		// team-a has a used configmap
		&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "shared-cm",
				Namespace: "team-a",
			},
			Data: map[string]string{
				"key": "value",
			},
		},
		// team-a pod uses shared-cm
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-a",
				Namespace: "team-a",
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "container-a",
						EnvFrom: []v1.EnvFromSource{
							{
								ConfigMapRef: &v1.ConfigMapEnvSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "shared-cm",
									},
								},
							},
						},
					},
				},
			},
		},
		// team-b has an unused configmap with the same name
		&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "shared-cm",
				Namespace: "team-b",
			},
			Data: map[string]string{
				"key": "value",
			},
		},
	)

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "",
	}

	analyzer := ConfigMapAnalyzer{}
	results, err := analyzer.Analyze(config)
	assert.NoError(t, err)

	// team-b/shared-cm is unused and must be reported; team-a/shared-cm is used and must not be reported.
	assert.Equal(t, 1, len(results), "expected 1 failure for the unused ConfigMap in team-b, got %d", len(results))
	if len(results) == 1 {
		assert.Equal(t, "team-b/shared-cm", results[0].Name)
		assert.Contains(t, results[0].Error[0].Text, "is not used by any pods in the namespace")
	}
}
