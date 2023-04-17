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

func TestHPAAnalyzerWithExistingScaleTargetRef(t *testing.T) {

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
