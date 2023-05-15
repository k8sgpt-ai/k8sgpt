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
	"strings"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/magiconair/properties/assert"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestHPAAnalyzer(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv1.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		})
	hpaAnalyzer := HpaAnalyzer{}
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	analysisResults, err := hpaAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 1)
}

func TestHPAAnalyzerWithMultipleHPA(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv1.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		},
		&autoscalingv1.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example-2",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		},
	)
	hpaAnalyzer := HpaAnalyzer{}
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	analysisResults, err := hpaAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 2)
}

func TestHPAAnalyzerWithUnsuportedScaleTargetRef(t *testing.T) {

	clientset := fake.NewSimpleClientset(
		&autoscalingv1.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
					Kind: "unsupported",
				},
			},
		})
	hpaAnalyzer := HpaAnalyzer{}

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	analysisResults, err := hpaAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}

	var errorFound bool
	for _, analysis := range analysisResults {
		for _, err := range analysis.Error {
			if strings.Contains(err.Text, "which is not an option.") {
				errorFound = true
				break
			}
		}
		if errorFound {
			break
		}
	}
	if !errorFound {
		t.Error("expected error 'does not possible option.' not found in analysis results")
	}
}

func TestHPAAnalyzerWithNonExistentScaleTargetRef(t *testing.T) {

	clientset := fake.NewSimpleClientset(
		&autoscalingv1.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
					Kind: "Deployment",
					Name: "non-existent",
				},
			},
		})
	hpaAnalyzer := HpaAnalyzer{}

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	analysisResults, err := hpaAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}

	var errorFound bool
	for _, analysis := range analysisResults {
		for _, err := range analysis.Error {
			if strings.Contains(err.Text, "does not exist.") {
				errorFound = true
				break
			}
		}
		if errorFound {
			break
		}
	}
	if !errorFound {
		t.Error("expected error 'does not exist.' not found in analysis results")
	}
}

func TestHPAAnalyzerWithExistingScaleTargetRefAsDeployment(t *testing.T) {

	clientset := fake.NewSimpleClientset(
		&autoscalingv1.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
					Kind: "Deployment",
					Name: "example",
				},
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "example",
								Image: "nginx",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"cpu":    resource.MustParse("100m"),
										"memory": resource.MustParse("128Mi"),
									},
									Limits: corev1.ResourceList{
										"cpu":    resource.MustParse("200m"),
										"memory": resource.MustParse("256Mi"),
									},
								},
							},
						},
					},
				},
			},
		},
	)
	hpaAnalyzer := HpaAnalyzer{}

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	analysisResults, err := hpaAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	for _, analysis := range analysisResults {
		assert.Equal(t, len(analysis.Error), 0)
	}
}

func TestHPAAnalyzerWithExistingScaleTargetRefAsReplicationController(t *testing.T) {

	clientset := fake.NewSimpleClientset(
		&autoscalingv1.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
					Kind: "ReplicationController",
					Name: "example",
				},
			},
		},
		&corev1.ReplicationController{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: corev1.ReplicationControllerSpec{
				Template: &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "example",
								Image: "nginx",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"cpu":    resource.MustParse("100m"),
										"memory": resource.MustParse("128Mi"),
									},
									Limits: corev1.ResourceList{
										"cpu":    resource.MustParse("200m"),
										"memory": resource.MustParse("256Mi"),
									},
								},
							},
						},
					},
				},
			},
		},
	)
	hpaAnalyzer := HpaAnalyzer{}

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	analysisResults, err := hpaAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	for _, analysis := range analysisResults {
		assert.Equal(t, len(analysis.Error), 0)
	}
}

func TestHPAAnalyzerWithExistingScaleTargetRefAsReplicaSet(t *testing.T) {

	clientset := fake.NewSimpleClientset(
		&autoscalingv1.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
					Kind: "ReplicaSet",
					Name: "example",
				},
			},
		},
		&appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: appsv1.ReplicaSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "example",
								Image: "nginx",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"cpu":    resource.MustParse("100m"),
										"memory": resource.MustParse("128Mi"),
									},
									Limits: corev1.ResourceList{
										"cpu":    resource.MustParse("200m"),
										"memory": resource.MustParse("256Mi"),
									},
								},
							},
						},
					},
				},
			},
		},
	)
	hpaAnalyzer := HpaAnalyzer{}

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	analysisResults, err := hpaAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	for _, analysis := range analysisResults {
		assert.Equal(t, len(analysis.Error), 0)
	}
}

func TestHPAAnalyzerWithExistingScaleTargetRefAsStatefulSet(t *testing.T) {

	clientset := fake.NewSimpleClientset(
		&autoscalingv1.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
					Kind: "StatefulSet",
					Name: "example",
				},
			},
		},
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: appsv1.StatefulSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "example",
								Image: "nginx",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"cpu":    resource.MustParse("100m"),
										"memory": resource.MustParse("128Mi"),
									},
									Limits: corev1.ResourceList{
										"cpu":    resource.MustParse("200m"),
										"memory": resource.MustParse("256Mi"),
									},
								},
							},
						},
					},
				},
			},
		},
	)
	hpaAnalyzer := HpaAnalyzer{}

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	analysisResults, err := hpaAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	for _, analysis := range analysisResults {
		assert.Equal(t, len(analysis.Error), 0)
	}
}

func TestHPAAnalyzerWithExistingScaleTargetRefWithoutSpecifyingResources(t *testing.T) {

	clientset := fake.NewSimpleClientset(
		&autoscalingv1.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
					Kind: "Deployment",
					Name: "example",
				},
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "example",
								Image: "nginx",
							},
						},
					},
				},
			},
		},
	)
	hpaAnalyzer := HpaAnalyzer{}

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	analysisResults, err := hpaAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}

	var errorFound bool
	for _, analysis := range analysisResults {
		for _, err := range analysis.Error {
			if strings.Contains(err.Text, "does not have resource configured.") {
				errorFound = true
				break
			}
			if errorFound {
				break
			}
		}
		if !errorFound {
			t.Error("expected error 'does not have resource configured.' not found in analysis results")
		}
	}
}

func TestHPAAnalyzerNamespaceFiltering(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv1.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		},
		&autoscalingv1.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "other-namespace",
				Annotations: map[string]string{},
			},
		})
	hpaAnalyzer := HpaAnalyzer{}
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	analysisResults, err := hpaAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 1)
}
