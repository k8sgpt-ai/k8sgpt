/*
Copyright 2026 The K8sGPT Authors.
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
	"errors"
	"fmt"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestHPAAnalyzerTargetLookupErrors(t *testing.T) {
	metadata := metav1.ObjectMeta{Name: "target", Namespace: "default"}
	targets := []struct {
		kind     string
		resource schema.GroupResource
		object   func(corev1.PodTemplateSpec) runtime.Object
	}{
		{
			kind: "Deployment", resource: appsv1.Resource("deployments"),
			object: func(template corev1.PodTemplateSpec) runtime.Object {
				return &appsv1.Deployment{ObjectMeta: metadata, Spec: appsv1.DeploymentSpec{Template: template}}
			},
		},
		{
			kind: "ReplicationController", resource: corev1.Resource("replicationcontrollers"),
			object: func(template corev1.PodTemplateSpec) runtime.Object {
				return &corev1.ReplicationController{ObjectMeta: metadata, Spec: corev1.ReplicationControllerSpec{Template: &template}}
			},
		},
		{
			kind: "ReplicaSet", resource: appsv1.Resource("replicasets"),
			object: func(template corev1.PodTemplateSpec) runtime.Object {
				return &appsv1.ReplicaSet{ObjectMeta: metadata, Spec: appsv1.ReplicaSetSpec{Template: template}}
			},
		},
		{
			kind: "StatefulSet", resource: appsv1.Resource("statefulsets"),
			object: func(template corev1.PodTemplateSpec) runtime.Object {
				return &appsv1.StatefulSet{ObjectMeta: metadata, Spec: appsv1.StatefulSetSpec{Template: template}}
			},
		},
	}
	lookups := []struct {
		name    string
		missing bool
		err     func(schema.GroupResource) error
	}{
		{name: "existing"},
		{name: "missing", missing: true},
		{name: "forbidden", err: func(resource schema.GroupResource) error {
			return apierrors.NewForbidden(resource, "target", errors.New("access denied"))
		}},
		{name: "timeout", err: func(schema.GroupResource) error {
			return apierrors.NewTimeoutError("API request timed out", 1)
		}},
		{name: "internal error", err: func(schema.GroupResource) error {
			return apierrors.NewInternalError(errors.New("API unavailable"))
		}},
		{name: "connection error", err: func(schema.GroupResource) error {
			return errors.New("connection reset by peer")
		}},
		{name: "wrapped not found", missing: true, err: func(resource schema.GroupResource) error {
			return fmt.Errorf("lookup failed: %w", apierrors.NewNotFound(resource, "target"))
		}},
	}

	for _, target := range targets {
		t.Run(target.kind, func(t *testing.T) {
			for _, lookup := range lookups {
				for _, unhealthy := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/unhealthy=%t", lookup.name, unhealthy), func(t *testing.T) {
						hpa := hpaLookupFixture("example", target.kind, "target", unhealthy)
						objects := []runtime.Object{hpa}
						if !lookup.missing || lookup.err != nil {
							objects = append(objects, target.object(hpaTargetTemplate(true)))
						}
						client := fake.NewSimpleClientset(objects...)
						if lookup.err != nil {
							client.PrependReactor("get", target.resource.Resource, func(action k8stesting.Action) (bool, runtime.Object, error) {
								require.Equal(t, "default", action.GetNamespace())
								require.Equal(t, "target", action.(k8stesting.GetAction).GetName())
								return true, nil, lookup.err(target.resource)
							})
						}

						var want []string
						if unhealthy {
							want = append(want, "scaling is unavailable")
						}
						if lookup.missing {
							want = append(want, fmt.Sprintf("HorizontalPodAutoscaler uses %s/target as ScaleTargetRef which does not exist.", target.kind))
						}
						got := hpaLookupFailures(t, client)
						if len(want) == 0 {
							require.Empty(t, got)
						} else {
							require.Len(t, got, 1)
							require.ElementsMatch(t, want, got["default/example"])
						}
					})
				}
			}

			t.Run("existing target without resources", func(t *testing.T) {
				client := fake.NewSimpleClientset(
					hpaLookupFixture("example", target.kind, "target", false),
					target.object(hpaTargetTemplate(false)),
				)
				require.Equal(t, []string{fmt.Sprintf("%s default/target does not have resource configured.", target.kind)}, hpaLookupFailures(t, client)["default/example"])
			})
		})
	}
}

func TestHPAAnalyzerTargetLookupDoesNotDiscardOtherFindings(t *testing.T) {
	client := fake.NewSimpleClientset(
		hpaLookupFixture("a-denied", "Deployment", "denied", true),
		hpaLookupFixture("b-missing", "Deployment", "missing", false),
		hpaLookupFixture("c-healthy", "Deployment", "healthy", false),
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "healthy", Namespace: "default"}, Spec: appsv1.DeploymentSpec{Template: hpaTargetTemplate(true)}},
	)
	client.PrependReactor("get", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if action.(k8stesting.GetAction).GetName() == "denied" {
			return true, nil, apierrors.NewForbidden(appsv1.Resource("deployments"), "denied", errors.New("access denied"))
		}
		return false, nil, nil
	})

	require.Equal(t, map[string][]string{
		"default/a-denied":  {"scaling is unavailable"},
		"default/b-missing": {"HorizontalPodAutoscaler uses Deployment/missing as ScaleTargetRef which does not exist."},
	}, hpaLookupFailures(t, client))
}

func TestHPAAnalyzerUnsupportedTargetDoesNotLookUpWorkloads(t *testing.T) {
	for _, unhealthy := range []bool{false, true} {
		t.Run(fmt.Sprintf("unhealthy=%t", unhealthy), func(t *testing.T) {
			client := fake.NewSimpleClientset(hpaLookupFixture("example", "CustomScale", "target", unhealthy))
			want := []string{"HorizontalPodAutoscaler uses CustomScale as ScaleTargetRef which is not an option."}
			if unhealthy {
				want = append(want, "scaling is unavailable")
			}
			require.ElementsMatch(t, want, hpaLookupFailures(t, client)["default/example"])
			for _, action := range client.Actions() {
				require.NotEqual(t, "get", action.GetVerb())
			}
		})
	}
}

func TestHPAAnalyzerTargetChecksPreserveListErrors(t *testing.T) {
	client := fake.NewSimpleClientset()
	wantErr := apierrors.NewForbidden(autoscalingv2.Resource("horizontalpodautoscalers"), "", errors.New("access denied"))
	client.PrependReactor("list", "horizontalpodautoscalers", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, wantErr
	})
	results, err := (HpaAnalyzer{}).Analyze(common.Analyzer{Client: &kubernetes.Client{Client: client}, Context: context.Background(), Namespace: "default"})
	require.ErrorIs(t, err, wantErr)
	require.Nil(t, results)
}

func hpaLookupFixture(name, kind, target string, unhealthy bool) *autoscalingv2.HorizontalPodAutoscaler {
	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{Kind: kind, Name: target},
		},
	}
	if unhealthy {
		hpa.Status.Conditions = []autoscalingv2.HorizontalPodAutoscalerCondition{{Type: autoscalingv2.AbleToScale, Status: corev1.ConditionFalse, Message: "scaling is unavailable"}}
	}
	return hpa
}

func hpaTargetTemplate(resources bool) corev1.PodTemplateSpec {
	container := corev1.Container{Name: "app", Image: "nginx"}
	if resources {
		container.Resources = corev1.ResourceRequirements{
			Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m")},
			Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("200m")},
		}
	}
	return corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{container}}}
}

func hpaLookupFailures(t *testing.T, client *fake.Clientset) map[string][]string {
	t.Helper()
	results, err := (HpaAnalyzer{}).Analyze(common.Analyzer{Client: &kubernetes.Client{Client: client}, Context: context.Background(), Namespace: "default"})
	require.NoError(t, err)
	failures := make(map[string][]string)
	for _, result := range results {
		require.Equal(t, "HorizontalPodAutoscaler", result.Kind)
		for _, failure := range result.Error {
			failures[result.Name] = append(failures[result.Name], failure.Text)
		}
	}
	return failures
}
