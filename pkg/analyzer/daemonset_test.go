package analyzer

import (
	"context"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/magiconair/properties/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestDaemonSetAnalyzer(t *testing.T) {
	clientset := fake.NewSimpleClientset(&appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "default",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"key": "value",
				},
			},
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "example-container",
							Image: "nginx",
						},
					},
				},
			},
		},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 1,
			NumberReady:            1,
		},
	})

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	daemonSetAnalyzer := DaemonSetAnalyzer{}
	analysisResults, err := daemonSetAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 0)
}

func TestDaemonSetAnalyzerNoReadyReplicas(t *testing.T) {
	clientset := fake.NewSimpleClientset(&appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "default",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"key": "value",
				},
			},
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
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 3,
			NumberReady:            2,
		},
	})

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	daemonSetAnalyzer := DaemonSetAnalyzer{}
	analysisResults, err := daemonSetAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 1)
	assert.Equal(t, analysisResults[0].Kind, "DaemonSet")
	assert.Equal(t, analysisResults[0].Name, "default/example")
}

func TestDaemonSetAnalyzerImagePullSecretNotFound(t *testing.T) {
	clientset := fake.NewSimpleClientset(&appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "default",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"key": "value",
				},
			},
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					ImagePullSecrets: []v1.LocalObjectReference{
						{
							Name: "non-existent-secret",
						},
					},
					Containers: []v1.Container{
						{
							Name:  "example-container",
							Image: "nginx",
						},
					},
				},
			},
		},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 1,
			NumberReady:            1,
		},
	})

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	daemonSetAnalyzer := DaemonSetAnalyzer{}
	analysisResults, err := daemonSetAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 1)
	assert.Equal(t, analysisResults[0].Kind, "DaemonSet")
	assert.Equal(t, analysisResults[0].Name, "default/example")
}

func newImagePullSecretDaemonSet() *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Name: "example", Namespace: "default"},
		Spec: appsv1.DaemonSetSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers:       []v1.Container{{Name: "c", Image: "nginx"}},
					ImagePullSecrets: []v1.LocalObjectReference{{Name: "regcred"}},
				},
			},
		},
		Status: appsv1.DaemonSetStatus{DesiredNumberScheduled: 1, NumberReady: 1},
	}
}

func analyzeDaemonSet(t *testing.T, clientset *fake.Clientset) []common.Result {
	t.Helper()
	results, err := DaemonSetAnalyzer{}.Analyze(common.Analyzer{
		Client:    &kubernetes.Client{Client: clientset},
		Context:   context.Background(),
		Namespace: "default",
	})
	if err != nil {
		t.Fatal(err)
	}
	return results
}

func TestDaemonSetAnalyzerExistingImagePullSecret(t *testing.T) {
	results := analyzeDaemonSet(t, fake.NewSimpleClientset(
		newImagePullSecretDaemonSet(),
		&v1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "regcred", Namespace: "default"}},
	))

	assert.Equal(t, len(results), 0)
}

// A secret that cannot be read (for example no RBAC on secrets) must not be
// reported as missing.
func TestDaemonSetAnalyzerImagePullSecretLookupForbidden(t *testing.T) {
	clientset := fake.NewSimpleClientset(newImagePullSecretDaemonSet())
	clientset.PrependReactor("get", "secrets", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, apierrors.NewForbidden(schema.GroupResource{Resource: "secrets"}, "regcred", nil)
	})

	assert.Equal(t, len(analyzeDaemonSet(t, clientset)), 0)
}
