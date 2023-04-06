package analyzer

import (
	"context"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/magiconair/properties/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestStatefulSetAnalyzer(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example",
				Namespace: "default",
			},
		})
	statefulSetAnalyzer := StatefulSetAnalyzer{}

	config := Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	analysisResults, err := statefulSetAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 1)
}

func TestStatefulSetAnalyzerWithoutService(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				ServiceName: "example-svc",
			},
		})
	statefulSetAnalyzer := StatefulSetAnalyzer{}

	config := Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	analysisResults, err := statefulSetAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	var errorFound bool
	want := "StatefulSet uses the service default/example-svc which does not exist."

	for _, analysis := range analysisResults {
		for _, got := range analysis.Error {
			if want == got {
				errorFound = true
			}
		}
		if errorFound {
			break
		}
	}
	if !errorFound {
		t.Errorf("Error expected: '%v', not found in StatefulSet's analysis results", want)
	}
}
