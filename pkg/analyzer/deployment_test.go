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
	"github.com/magiconair/properties/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDeploymentAnalyzer(t *testing.T) {
	clientset := fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func() *int32 { i := int32(3); return &i }(),
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "example-container",
							Image: "nginx",
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			Replicas:          2,
			AvailableReplicas: 1,
		},
	})

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	deploymentAnalyzer := DeploymentAnalyzer{}
	analysisResults, err := deploymentAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 1)
	assert.Equal(t, analysisResults[0].Kind, "Deployment")
	assert.Equal(t, analysisResults[0].Name, "default/example")
}

func TestDeploymentAnalyzerNamespaceFiltering(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: func() *int32 { i := int32(3); return &i }(),
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "example-container",
								Image: "nginx",
								Ports: []v1.ContainerPort{
									{
										ContainerPort: 80,
									},
								},
							},
						},
					},
				},
			},
			Status: appsv1.DeploymentStatus{
				Replicas:          2,
				AvailableReplicas: 1,
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example",
				Namespace: "other-namespace",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: func() *int32 { i := int32(3); return &i }(),
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "example-container",
								Image: "nginx",
								Ports: []v1.ContainerPort{
									{
										ContainerPort: 80,
									},
								},
							},
						},
					},
				},
			},
			Status: appsv1.DeploymentStatus{
				Replicas:          2,
				AvailableReplicas: 1,
			},
		},
	)

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	deploymentAnalyzer := DeploymentAnalyzer{}
	analysisResults, err := deploymentAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 1)
	assert.Equal(t, analysisResults[0].Kind, "Deployment")
	assert.Equal(t, analysisResults[0].Name, "default/example")
}

func TestDeploymentAnalyzerLabelSelectorFiltering(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example",
				Namespace: "default",
				Labels: map[string]string{
					"app": "deployment",
				},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: func() *int32 { i := int32(3); return &i }(),
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{},
					},
				},
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example2",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: func() *int32 { i := int32(3); return &i }(),
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{},
					},
				},
			},
		},
	)

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:       context.Background(),
		Namespace:     "default",
		LabelSelector: "app=deployment",
	}

	deploymentAnalyzer := DeploymentAnalyzer{}
	analysisResults, err := deploymentAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 1)
}
