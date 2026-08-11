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

// TestConfigMapAnalyzer_InitContainerUsage reproduces the false positive where a
// ConfigMap referenced only from an init container's env/envFrom is reported as unused.
func TestConfigMapAnalyzer_InitContainerUsage(t *testing.T) {
	tests := []struct {
		name           string
		configMapName  string
		pod            v1.Pod
		expectedErrors int
	}{
		{
			name:          "configmap used by init container via envFrom",
			configMapName: "init-cm",
			pod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "init-pod", Namespace: "default"},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name: "init",
							EnvFrom: []v1.EnvFromSource{
								{ConfigMapRef: &v1.ConfigMapEnvSource{LocalObjectReference: v1.LocalObjectReference{Name: "init-cm"}}},
							},
						},
					},
					Containers: []v1.Container{{Name: "app"}},
				},
			},
			expectedErrors: 0,
		},
		{
			name:          "configmap used by init container via env valueFrom",
			configMapName: "init-env-cm",
			pod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "init-pod2", Namespace: "default"},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name: "init",
							Env: []v1.EnvVar{
								{Name: "FOO", ValueFrom: &v1.EnvVarSource{ConfigMapKeyRef: &v1.ConfigMapKeySelector{LocalObjectReference: v1.LocalObjectReference{Name: "init-env-cm"}, Key: "foo"}}},
							},
						},
					},
					Containers: []v1.Container{{Name: "app"}},
				},
			},
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			_, err := client.CoreV1().ConfigMaps("default").Create(context.TODO(), &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: tt.configMapName, Namespace: "default"},
				Data:       map[string]string{"foo": "bar"},
			}, metav1.CreateOptions{})
			assert.NoError(t, err)
			_, err = client.CoreV1().Pods("default").Create(context.TODO(), &tt.pod, metav1.CreateOptions{})
			assert.NoError(t, err)

			analyzer := ConfigMapAnalyzer{}
			results, err := analyzer.Analyze(common.Analyzer{
				Client:    &kubernetes.Client{Client: client},
				Context:   context.TODO(),
				Namespace: "default",
			})
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedErrors, len(results), "expected %d errors but got %d", tt.expectedErrors, len(results))
		})
	}
}
