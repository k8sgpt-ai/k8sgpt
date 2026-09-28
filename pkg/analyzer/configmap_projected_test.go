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

// TestConfigMapAnalyzer_ProjectedVolumeUsage reproduces the false positive where a
// ConfigMap referenced only from a projected volume source is reported as unused.
func TestConfigMapAnalyzer_ProjectedVolumeUsage(t *testing.T) {
	tests := []struct {
		name           string
		configMapName  string
		pod            v1.Pod
		expectedErrors int
	}{
		{
			name:          "configmap used only via projected volume source",
			configMapName: "projected-cm",
			pod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "projected-pod", Namespace: "default"},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: "cfg",
							VolumeSource: v1.VolumeSource{
								Projected: &v1.ProjectedVolumeSource{
									Sources: []v1.VolumeProjection{
										{
											ConfigMap: &v1.ConfigMapProjection{
												LocalObjectReference: v1.LocalObjectReference{Name: "projected-cm"},
											},
										},
									},
								},
							},
						},
					},
					Containers: []v1.Container{{Name: "app"}},
				},
			},
			expectedErrors: 0,
		},
		{
			name:          "kube-root-ca.crt used via kubelet projected service-account volume",
			configMapName: "kube-root-ca.crt",
			pod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "sa-pod", Namespace: "default"},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							Name: "kube-api-access",
							VolumeSource: v1.VolumeSource{
								Projected: &v1.ProjectedVolumeSource{
									Sources: []v1.VolumeProjection{
										{
											ServiceAccountToken: &v1.ServiceAccountTokenProjection{
												Path: "token",
											},
										},
										{
											ConfigMap: &v1.ConfigMapProjection{
												LocalObjectReference: v1.LocalObjectReference{Name: "kube-root-ca.crt"},
												Items: []v1.KeyToPath{
													{Key: "ca.crt", Path: "ca.crt"},
												},
											},
										},
										{
											DownwardAPI: &v1.DownwardAPIProjection{
												Items: []v1.DownwardAPIVolumeFile{
													{Path: "namespace"},
												},
											},
										},
									},
								},
							},
						},
					},
					Containers: []v1.Container{{Name: "app"}},
				},
			},
			expectedErrors: 0,
		},
		{
			name:          "unused configmap is still reported",
			configMapName: "unused-cm",
			pod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "other-pod", Namespace: "default"},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Name: "app"}},
				},
			},
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			_, err := client.CoreV1().ConfigMaps("default").Create(context.TODO(), &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: tt.configMapName, Namespace: "default"},
				Data:       map[string]string{"ca.crt": "dummy", "foo": "bar"},
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
