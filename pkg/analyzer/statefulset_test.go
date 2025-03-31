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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

	config := common.Analyzer{
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

	config := common.Analyzer{
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
			if want == got.Text {
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

func TestStatefulSetAnalyzerMissingStorageClass(t *testing.T) {
	storageClassName := "example-sc"
	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				ServiceName: "example-svc",
				VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
					{
						TypeMeta: metav1.TypeMeta{
							Kind:       "PersistentVolumeClaim",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name: "pvc-example",
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							StorageClassName: &storageClassName,
							AccessModes: []corev1.PersistentVolumeAccessMode{
								"ReadWriteOnce",
							},
							Resources: corev1.VolumeResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceStorage: resource.MustParse("1Gi"),
								},
							},
						},
					},
				},
			},
		})
	statefulSetAnalyzer := StatefulSetAnalyzer{}

	config := common.Analyzer{
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
	want := "StatefulSet uses the storage class example-sc which does not exist."

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
		t.Errorf("Error expected: '%v', not found in StatefulSet's analysis results", want)
	}

}

func TestStatefulSetAnalyzerNamespaceFiltering(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example",
				Namespace: "default",
			},
		},
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example",
				Namespace: "other-namespace",
			},
		})
	statefulSetAnalyzer := StatefulSetAnalyzer{}

	config := common.Analyzer{
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

func TestStatefulSetAnalyzerLabelSelectorFiltering(t *testing.T) {
	clientSet := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example1",
				Namespace: "default",
				Labels: map[string]string{
					"app":     "statefulset",
					"part-of": "test",
				},
			},
		},
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example2",
				Namespace: "default",
			},
		},
	)
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientSet,
		},
		Context:       context.Background(),
		Namespace:     "default",
		LabelSelector: "app=statefulset",
	}
	statefulSetAnalyzer := StatefulSetAnalyzer{}
	results, err := statefulSetAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "default/example1", results[0].Name)

	config = common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientSet,
		},
		Context:       context.Background(),
		Namespace:     "default",
		LabelSelector: "app=statefulset,part-of=test",
	}
	statefulSetAnalyzer = StatefulSetAnalyzer{}
	results, err = statefulSetAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "default/example1", results[0].Name)
}

func TestStatefulSetAnalyzerReplica(t *testing.T) {
	replicas := int32(3)
	pods := []*corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example-0",
				Namespace: "default",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example-1",
				Namespace: "default",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example-2",
				Namespace: "default",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
	}
	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: &replicas,
			},
			Status: appsv1.StatefulSetStatus{
				AvailableReplicas: 3,
			},
		},
		pods[0], pods[1], pods[2],
	)
	statefulSetAnalyzer := StatefulSetAnalyzer{}

	config := common.Analyzer{
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

func TestStatefulSetAnalyzerUnavailableReplicas(t *testing.T) {
	replicas := int32(3)
	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: &replicas,
			},
			Status: appsv1.StatefulSetStatus{
				AvailableReplicas: 0,
			},
		})
	statefulSetAnalyzer := StatefulSetAnalyzer{}

	config := common.Analyzer{
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

func TestStatefulSetAnalyzerUnavailableReplicaWithPodInitialized(t *testing.T) {
	replicas := int32(3)
	pods := []*corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example-0",
				Namespace: "default",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example-1",
				Namespace: "default",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
			},
		},
	}
	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: &replicas,
			},
			Status: appsv1.StatefulSetStatus{
				AvailableReplicas: 1,
			},
		},
		pods[0], pods[1],
	)
	statefulSetAnalyzer := StatefulSetAnalyzer{}

	config := common.Analyzer{
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
	want := "Statefulset pod example-1 in the namespace default is not in running state."

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
		t.Errorf("Error expected: '%v', not found in StatefulSet's analysis results", want)
	}
}
