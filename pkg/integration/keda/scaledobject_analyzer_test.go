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

package keda

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	kedaSchema "github.com/kedacore/keda/v2/apis/keda/v1alpha1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

// newScaledObjectServer returns an httptest server that serves the given
// ScaledObjects from the KEDA API. The analyzer builds its KEDA client from
// the rest.Config, so a real HTTP endpoint is needed to drive List.
func newScaledObjectServer(t *testing.T, items []kedaSchema.ScaledObject) *httptest.Server {
	t.Helper()
	list := kedaSchema.ScaledObjectList{
		TypeMeta: metav1.TypeMeta{Kind: "ScaledObjectList", APIVersion: "keda.sh/v1alpha1"},
		Items:    items,
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(list); err != nil {
			t.Errorf("encoding ScaledObjectList: %v", err)
		}
	}))
}

func cpuScaledObject(name, namespace string) kedaSchema.ScaledObject {
	return kedaSchema.ScaledObject{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: kedaSchema.ScaledObjectSpec{
			ScaleTargetRef: &kedaSchema.ScaleTarget{Name: name, Kind: "Deployment"},
			Triggers:       []kedaSchema.ScaleTriggers{{Type: "cpu"}},
		},
	}
}

// TestScaledObjectAnalyzerReportsResourceFailureWithoutEvent is the regression
// for the case where a ScaledObject has no associated event. Previously the
// analyzer did an unconditional `continue` when FetchLatestEvent returned a nil
// event, which skipped the failure-commit block and silently dropped the
// "does not have resource configured." finding collected earlier in the loop.
func TestScaledObjectAnalyzerReportsResourceFailureWithoutEvent(t *testing.T) {
	srv := newScaledObjectServer(t, []kedaSchema.ScaledObject{cpuScaledObject("example", "default")})
	defer srv.Close()

	// Target Deployment exists but its container declares no resource
	// requests/limits, so a cpu-triggered ScaledObject is misconfigured. No
	// event exists for the ScaledObject (the common case).
	clientset := fake.NewSimpleClientset(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "example", Namespace: "default"},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "app"}},
					},
				},
			},
		},
	)

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
			Config: &rest.Config{Host: srv.URL},
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	results, err := (&ScaledObjectAnalyzer{}).Analyze(config)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "ScaledObject", results[0].Kind)
	assert.Equal(t, "default/example", results[0].Name)
	assert.Equal(t, "Deployment default/example does not have resource configured.", results[0].Error[0].Text)
}

// TestScaledObjectAnalyzerReportsMissingTargetWithoutEvent verifies the
// ScaleTargetRef-does-not-exist failure is also reported when there is no
// event (another path that the previous early-continue discarded).
func TestScaledObjectAnalyzerReportsMissingTargetWithoutEvent(t *testing.T) {
	srv := newScaledObjectServer(t, []kedaSchema.ScaledObject{cpuScaledObject("missing", "default")})
	defer srv.Close()

	// No Deployment named "missing" exists, and no event exists either.
	clientset := fake.NewSimpleClientset()

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
			Config: &rest.Config{Host: srv.URL},
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	results, err := (&ScaledObjectAnalyzer{}).Analyze(config)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "ScaledObject uses Deployment/missing as ScaleTargetRef which does not exist.", results[0].Error[0].Text)
}

// TestScaledObjectAnalyzerReportsWarningEvent verifies the event-based failure
// path still works: a non-Normal event message is surfaced.
func TestScaledObjectAnalyzerReportsWarningEvent(t *testing.T) {
	srv := newScaledObjectServer(t, []kedaSchema.ScaledObject{cpuScaledObject("example", "default")})
	defer srv.Close()

	// Target is correctly configured (requests and limits set), so the only
	// failure comes from the Warning event attached to the ScaledObject.
	clientset := fake.NewSimpleClientset(
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "example", Namespace: "default"},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "app",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m")},
									Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("200m")},
								},
							},
						},
					},
				},
			},
		},
		&corev1.Event{
			ObjectMeta:     metav1.ObjectMeta{Name: "example.event", Namespace: "default"},
			InvolvedObject: corev1.ObjectReference{Name: "example", Namespace: "default"},
			LastTimestamp:  metav1.Now(),
			Type:           "Warning",
			Message:        "ScaledObject doesn't have correct scaleTargetRef specification",
		},
	)

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
			Config: &rest.Config{Host: srv.URL},
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	results, err := (&ScaledObjectAnalyzer{}).Analyze(config)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "ScaledObject doesn't have correct scaleTargetRef specification", results[0].Error[0].Text)
}
