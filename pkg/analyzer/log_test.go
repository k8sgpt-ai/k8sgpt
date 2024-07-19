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
	"regexp"
	"sort"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLogAnalyzer(t *testing.T) {
	oldPattern := errorPattern
	errorPattern = regexp.MustCompile(`(fake logs)`)
	t.Cleanup(func() {
		errorPattern = oldPattern
	})

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: fake.NewSimpleClientset(
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Pod1",
						Namespace: "default",
						Labels: map[string]string{
							"Name":      "Pod1",
							"Namespace": "default",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "test-container1",
							},
							{
								Name: "test-container2",
							},
						},
					},
				},
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Pod2",
						Namespace: "default",
						Labels: map[string]string{
							"Name":      "Pod1",
							"Namespace": "default",
						},
					},
				},
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Pod3",
						Namespace: "test-namespace",
						Labels: map[string]string{
							"Name":      "Pod1",
							"Namespace": "test-namespace",
						},
					},
				},
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Pod4",
						Namespace: "default",
						Labels: map[string]string{
							"Name":      "Pod4",
							"Namespace": "default",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "test-container3",
							},
						},
					},
				},
			),
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	logAnalyzer := LogAnalyzer{}
	results, err := logAnalyzer.Analyze(config)
	require.NoError(t, err)

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	expectations := []string{"default/Pod1/test-container1", "default/Pod1/test-container2", "default/Pod4/test-container3"}

	for i, expectation := range expectations {
		require.Equal(t, expectation, results[i].Name)

		for _, failure := range results[i].Error {
			require.Equal(t, "fake logs", failure.Text)
		}
	}
}

func TestLogAnalyzerLabelSelectorFiltering(t *testing.T) {
	oldPattern := errorPattern
	errorPattern = regexp.MustCompile(`(fake logs)`)
	t.Cleanup(func() {
		errorPattern = oldPattern
	})

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: fake.NewSimpleClientset(
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Pod1",
						Namespace: "default",
						Labels: map[string]string{
							"app": "log",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "test-container1",
							},
						},
					},
				},
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Pod2",
						Namespace: "default",
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "test-container2",
							},
						},
					},
				},
			),
		},
		Context:       context.Background(),
		Namespace:     "default",
		LabelSelector: "app=log",
	}

	logAnalyzer := LogAnalyzer{}
	results, err := logAnalyzer.Analyze(config)
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, "default/Pod1/test-container1", results[0].Name)
}
