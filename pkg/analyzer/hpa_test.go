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
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestHPAAnalyzer(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv2.HorizontalPodAutoscaler{
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
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		},
		&autoscalingv2.HorizontalPodAutoscaler{
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
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
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
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
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
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
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
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
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
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
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
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
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
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
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
			if strings.Contains(err.Text, "does not set a cpu request on container(s) example") {
				errorFound = true
				break
			}
		}
		if errorFound {
			break
		}
	}
	if !errorFound {
		t.Error("expected error 'does not set a cpu request on container(s) example' not found in analysis results")
	}
}

func TestHPAAnalyzerNamespaceFiltering(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		},
		&autoscalingv2.HorizontalPodAutoscaler{
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

func TestHPAAnalyzerLabelSelectorFiltering(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example",
				Namespace: "default",
				Labels: map[string]string{
					"app": "hpa",
				},
			},
		},
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example2",
				Namespace: "default",
			},
		},
	)
	hpaAnalyzer := HpaAnalyzer{}
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:       context.Background(),
		Namespace:     "default",
		LabelSelector: "app=hpa",
	}
	analysisResults, err := hpaAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 1)
}

func TestHPAAnalyzerStatusFieldAbleToScale(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
					Kind: "Deployment",
					Name: "example",
				},
			},
			Status: autoscalingv2.HorizontalPodAutoscalerStatus{
				Conditions: []autoscalingv2.HorizontalPodAutoscalerCondition{
					{
						Type:    "AbleToScale",
						Status:  "False",
						Message: "test reason",
					},
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
	assert.Equal(t, len(analysisResults), 1)

}

func TestHPAAnalyzerStatusFieldScalingActive(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
					Kind: "Deployment",
					Name: "example",
				},
			},
			Status: autoscalingv2.HorizontalPodAutoscalerStatus{
				Conditions: []autoscalingv2.HorizontalPodAutoscalerCondition{
					{
						Type:    autoscalingv2.ScalingActive,
						Status:  "False",
						Message: "test reason",
					},
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
	assert.Equal(t, len(analysisResults), 1)

}

func TestHPAAnalyzerStatusFieldScalingLimited(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
					Kind: "Deployment",
					Name: "example",
				},
			},
			Status: autoscalingv2.HorizontalPodAutoscalerStatus{
				Conditions: []autoscalingv2.HorizontalPodAutoscalerCondition{
					{
						Type:    autoscalingv2.ScalingLimited,
						Status:  "False",
						Message: "test reason",
					},
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
	assert.Equal(t, len(analysisResults), 1)

}

func TestHPAAnalyzerStatusField(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
					Kind: "Deployment",
					Name: "example",
				},
			},
			Status: autoscalingv2.HorizontalPodAutoscalerStatus{
				Conditions: []autoscalingv2.HorizontalPodAutoscalerCondition{
					{
						Type:    autoscalingv2.AbleToScale,
						Status:  "True",
						Message: "recommended size matches current size",
					},
					{
						Type:    autoscalingv2.ScalingActive,
						Status:  "True",
						Message: "the HPA was able to successfully calculate a replica count",
					},
					{
						Type:    autoscalingv2.ScalingLimited,
						Status:  "True",
						Message: "the desired replica count is less than the minimum replica count",
					},
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
	assert.Equal(t, len(analysisResults), 1)

}

func TestHPAAnalyzerStatusScalingLimitedError(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
					Kind: "Deployment",
					Name: "example",
				},
			},
			Status: autoscalingv2.HorizontalPodAutoscalerStatus{
				Conditions: []autoscalingv2.HorizontalPodAutoscalerCondition{
					{
						Type:    autoscalingv2.AbleToScale,
						Status:  "True",
						Message: "recommended size matches current size",
					},
					{
						Type:    autoscalingv2.ScalingActive,
						Status:  "True",
						Message: "the HPA was able to successfully calculate a replica count",
					},
					{
						Type:    autoscalingv2.ScalingLimited,
						Status:  "True",
						Message: "the desired replica count is less than the minimum replica count",
					},
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
	want := "the desired replica count is less than the minimum replica count"
	for _, analysis := range analysisResults {
		for _, got := range analysis.Error {
			if want == got.Text {
				errorFound = true
			}
		}
		if errorFound {
			break
		}
	}

	if !errorFound {
		t.Errorf("Expected message, <%v> , not found in HorizontalPodAutoscaler's analysis results", want)
	}
}

func TestHPAAnalyzerScalingLimitedTooFewReplicasAtMinIgnored(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
					Kind: "Deployment",
					Name: "example",
				},
				MinReplicas: func() *int32 { i := int32(2); return &i }(),
				MaxReplicas: 6,
			},
			Status: autoscalingv2.HorizontalPodAutoscalerStatus{
				CurrentReplicas: 2,
				DesiredReplicas: 2,
				Conditions: []autoscalingv2.HorizontalPodAutoscalerCondition{
					{
						Type:    autoscalingv2.ScalingLimited,
						Status:  corev1.ConditionTrue,
						Reason:  "TooFewReplicas",
						Message: "the desired replica count is less than the minimum replica count",
					},
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
	assert.Equal(t, len(analysisResults), 0)
}

func TestHPAAnalyzerScalingLimitedTooFewReplicasAboveMinReported(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
					Kind: "Deployment",
					Name: "example",
				},
				MinReplicas: func() *int32 { i := int32(2); return &i }(),
				MaxReplicas: 6,
			},
			Status: autoscalingv2.HorizontalPodAutoscalerStatus{
				CurrentReplicas: 5,
				DesiredReplicas: 5,
				Conditions: []autoscalingv2.HorizontalPodAutoscalerCondition{
					{
						Type:    autoscalingv2.ScalingLimited,
						Status:  corev1.ConditionTrue,
						Reason:  "TooFewReplicas",
						Message: "the desired replica count is less than the minimum replica count",
					},
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
	assert.Equal(t, len(analysisResults), 1)
	assert.Equal(t, len(analysisResults[0].Error), 1)
	assert.Equal(t, analysisResults[0].Error[0].Text, "the desired replica count is less than the minimum replica count")
}

func TestHPAAnalyzerScalingLimitedTooManyReplicasReported(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
					Kind: "Deployment",
					Name: "example",
				},
				MinReplicas: func() *int32 { i := int32(2); return &i }(),
				MaxReplicas: 2,
			},
			Status: autoscalingv2.HorizontalPodAutoscalerStatus{
				CurrentReplicas: 2,
				DesiredReplicas: 2,
				Conditions: []autoscalingv2.HorizontalPodAutoscalerCondition{
					{
						Type:    autoscalingv2.ScalingLimited,
						Status:  corev1.ConditionTrue,
						Reason:  "TooManyReplicas",
						Message: "the desired replica count is more than the maximum replica count",
					},
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
	assert.Equal(t, len(analysisResults), 1)
	assert.Equal(t, len(analysisResults[0].Error), 1)
	assert.Equal(t, analysisResults[0].Error[0].Text, "the desired replica count is more than the maximum replica count")
}

func TestHPAAnalyzerUnsupportedScaleTargetRefReportedOnce(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example",
				Namespace: "default",
			},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
					Kind: "unsupported",
					Name: "foo",
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

	if len(analysisResults) != 1 {
		t.Fatalf("expected 1 result, got %d", len(analysisResults))
	}

	// an unsupported kind is never looked up, so it must not also be reported
	// as a target that does not exist
	failures := analysisResults[0].Error
	if len(failures) != 1 {
		var texts []string
		for _, f := range failures {
			texts = append(texts, f.Text)
		}
		t.Fatalf("expected 1 failure, got %d: %v", len(failures), texts)
	}
	if !strings.Contains(failures[0].Text, "which is not an option.") {
		t.Errorf("unexpected failure text: %s", failures[0].Text)
	}
}

func TestHPAAnalyzerUtilizationRequests(t *testing.T) {
	always := corev1.ContainerRestartPolicyAlways
	utilization := int32(80)

	resourceUtilization := func(name corev1.ResourceName) autoscalingv2.MetricSpec {
		return autoscalingv2.MetricSpec{
			Type: autoscalingv2.ResourceMetricSourceType,
			Resource: &autoscalingv2.ResourceMetricSource{
				Name: name,
				Target: autoscalingv2.MetricTarget{
					Type:               autoscalingv2.UtilizationMetricType,
					AverageUtilization: &utilization,
				},
			},
		}
	}
	containerUtilization := func(name corev1.ResourceName, container string) autoscalingv2.MetricSpec {
		return autoscalingv2.MetricSpec{
			Type: autoscalingv2.ContainerResourceMetricSourceType,
			ContainerResource: &autoscalingv2.ContainerResourceMetricSource{
				Name:      name,
				Container: container,
				Target: autoscalingv2.MetricTarget{
					Type:               autoscalingv2.UtilizationMetricType,
					AverageUtilization: &utilization,
				},
			},
		}
	}
	requests := func(names ...corev1.ResourceName) corev1.ResourceRequirements {
		list := corev1.ResourceList{}
		for _, name := range names {
			list[name] = resource.MustParse("100m")
		}
		return corev1.ResourceRequirements{Requests: list}
	}
	limits := func(names ...corev1.ResourceName) corev1.ResourceRequirements {
		return corev1.ResourceRequirements{Limits: requests(names...).Requests}
	}
	averageValue := resource.MustParse("200m")

	tests := []struct {
		name     string
		metrics  []autoscalingv2.MetricSpec
		podSpec  corev1.PodSpec
		expected []string
	}{
		{
			name:    "requests without limits are enough",
			metrics: []autoscalingv2.MetricSpec{resourceUtilization(corev1.ResourceCPU)},
			podSpec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "app", Resources: requests(corev1.ResourceCPU)}},
			},
		},
		{
			name:    "limits without requests default the requests",
			metrics: []autoscalingv2.MetricSpec{resourceUtilization(corev1.ResourceCPU)},
			podSpec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "app", Resources: limits(corev1.ResourceCPU)}},
			},
		},
		{
			name: "no metrics defaults to cpu utilization",
			podSpec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "app", Resources: requests(corev1.ResourceMemory)}},
			},
			expected: []string{
				"Deployment default/example does not set a cpu request on container(s) app, so the HorizontalPodAutoscaler cannot compute its cpu utilization.",
			},
		},
		{
			name:    "one container without a request",
			metrics: []autoscalingv2.MetricSpec{resourceUtilization(corev1.ResourceCPU)},
			podSpec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "app", Resources: requests(corev1.ResourceCPU)},
					{Name: "sidecar"},
				},
			},
			expected: []string{
				"Deployment default/example does not set a cpu request on container(s) sidecar, so the HorizontalPodAutoscaler cannot compute its cpu utilization.",
			},
		},
		{
			name:    "memory utilization needs a memory request",
			metrics: []autoscalingv2.MetricSpec{resourceUtilization(corev1.ResourceMemory)},
			podSpec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "app", Resources: requests(corev1.ResourceCPU)}},
			},
			expected: []string{
				"Deployment default/example does not set a memory request on container(s) app, so the HorizontalPodAutoscaler cannot compute its memory utilization.",
			},
		},
		{
			name: "average value target does not need requests",
			metrics: []autoscalingv2.MetricSpec{{
				Type: autoscalingv2.ResourceMetricSourceType,
				Resource: &autoscalingv2.ResourceMetricSource{
					Name: corev1.ResourceCPU,
					Target: autoscalingv2.MetricTarget{
						Type:         autoscalingv2.AverageValueMetricType,
						AverageValue: &averageValue,
					},
				},
			}},
			podSpec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}}},
		},
		{
			name: "pods metric does not need requests",
			metrics: []autoscalingv2.MetricSpec{{
				Type: autoscalingv2.PodsMetricSourceType,
				Pods: &autoscalingv2.PodsMetricSource{
					Metric: autoscalingv2.MetricIdentifier{Name: "requests_per_second"},
					Target: autoscalingv2.MetricTarget{
						Type:         autoscalingv2.AverageValueMetricType,
						AverageValue: &averageValue,
					},
				},
			}},
			podSpec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}}},
		},
		{
			name:    "sidecar init container needs a request",
			metrics: []autoscalingv2.MetricSpec{resourceUtilization(corev1.ResourceCPU)},
			podSpec: corev1.PodSpec{
				InitContainers: []corev1.Container{
					{Name: "migrate"},
					{Name: "proxy", RestartPolicy: &always},
				},
				Containers: []corev1.Container{{Name: "app", Resources: requests(corev1.ResourceCPU)}},
			},
			expected: []string{
				"Deployment default/example does not set a cpu request on container(s) proxy, so the HorizontalPodAutoscaler cannot compute its cpu utilization.",
			},
		},
		{
			name:    "pod-level request covers every container",
			metrics: []autoscalingv2.MetricSpec{resourceUtilization(corev1.ResourceCPU)},
			podSpec: corev1.PodSpec{
				Resources:  &corev1.ResourceRequirements{Requests: requests(corev1.ResourceCPU).Requests},
				Containers: []corev1.Container{{Name: "app"}, {Name: "sidecar"}},
			},
		},
		{
			name:    "pod-level limit defaults the pod-level request",
			metrics: []autoscalingv2.MetricSpec{resourceUtilization(corev1.ResourceCPU)},
			podSpec: corev1.PodSpec{
				Resources:  &corev1.ResourceRequirements{Limits: limits(corev1.ResourceCPU).Limits},
				Containers: []corev1.Container{{Name: "app"}},
			},
		},
		{
			name:    "pod-level request falls back to the container requests",
			metrics: []autoscalingv2.MetricSpec{resourceUtilization(corev1.ResourceCPU)},
			podSpec: corev1.PodSpec{
				Resources: &corev1.ResourceRequirements{Requests: requests(corev1.ResourceMemory).Requests},
				Containers: []corev1.Container{
					{Name: "app", Resources: requests(corev1.ResourceCPU)},
					{Name: "sidecar"},
				},
			},
		},
		{
			name:    "pod-level request without the resource anywhere",
			metrics: []autoscalingv2.MetricSpec{resourceUtilization(corev1.ResourceCPU)},
			podSpec: corev1.PodSpec{
				Resources:  &corev1.ResourceRequirements{Requests: requests(corev1.ResourceMemory).Requests},
				Containers: []corev1.Container{{Name: "app"}, {Name: "sidecar"}},
			},
			expected: []string{
				"Deployment default/example does not set a cpu request on container(s) app, sidecar, so the HorizontalPodAutoscaler cannot compute its cpu utilization.",
			},
		},
		{
			name: "container resource metric ignores the pod-level request",
			metrics: []autoscalingv2.MetricSpec{
				containerUtilization(corev1.ResourceCPU, "app"),
			},
			podSpec: corev1.PodSpec{
				Resources:  &corev1.ResourceRequirements{Requests: requests(corev1.ResourceCPU).Requests},
				Containers: []corev1.Container{{Name: "app"}},
			},
			expected: []string{
				"Deployment default/example does not set a cpu request on container(s) app, so the HorizontalPodAutoscaler cannot compute its cpu utilization.",
			},
		},
		{
			name:    "container resource metric only needs the named container",
			metrics: []autoscalingv2.MetricSpec{containerUtilization(corev1.ResourceCPU, "app")},
			podSpec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "app", Resources: requests(corev1.ResourceCPU)},
					{Name: "sidecar"},
				},
			},
		},
		{
			name:    "container resource metric on a missing container",
			metrics: []autoscalingv2.MetricSpec{containerUtilization(corev1.ResourceCPU, "web")},
			podSpec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "app", Resources: requests(corev1.ResourceCPU)}},
			},
			expected: []string{
				"HorizontalPodAutoscaler scales on the cpu utilization of container web, which does not exist in Deployment default/example.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset(
				&autoscalingv2.HorizontalPodAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "example",
						Namespace: "default",
					},
					Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
						ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
							Kind: "Deployment",
							Name: "example",
						},
						Metrics: tt.metrics,
					},
				},
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "example",
						Namespace: "default",
					},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{Spec: tt.podSpec},
					},
				},
			)

			// analyze every namespace so the failure text has to name the
			// namespace of the HPA rather than the one being analyzed
			config := common.Analyzer{
				Client: &kubernetes.Client{
					Client: clientset,
				},
				Context: context.Background(),
			}

			analysisResults, err := HpaAnalyzer{}.Analyze(config)
			if err != nil {
				t.Fatal(err)
			}

			var texts []string
			for _, result := range analysisResults {
				for _, failure := range result.Error {
					texts = append(texts, failure.Text)
				}
			}
			if len(texts) != len(tt.expected) {
				t.Fatalf("expected failures %q, got %q", tt.expected, texts)
			}
			for i := range texts {
				if texts[i] != tt.expected[i] {
					t.Errorf("expected failure %q, got %q", tt.expected[i], texts[i])
				}
			}
		})
	}
}
