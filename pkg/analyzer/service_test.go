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
	"sort"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

func TestServiceAnalyzer(t *testing.T) {
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: fake.NewSimpleClientset(
				&v1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Endpoint1",
						Namespace: "test",
					},
					// Endpoint with non-zero subsets.
					Subsets: []v1.EndpointSubset{
						{
							// These not ready end points will contribute to failures.
							NotReadyAddresses: []v1.EndpointAddress{
								{
									TargetRef: &v1.ObjectReference{
										Kind: "test-reference",
										Name: "reference1",
									},
								},
								{
									TargetRef: &v1.ObjectReference{
										Kind: "test-reference",
										Name: "reference2",
									},
								},
							},
						},
						{
							// These not ready end points will contribute to failures.
							NotReadyAddresses: []v1.EndpointAddress{
								{
									TargetRef: &v1.ObjectReference{
										Kind: "test-reference",
										Name: "reference3",
									},
								},
							},
						},
					},
				},
				&v1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Endpoint2",
						Namespace: "test",
						Annotations: map[string]string{
							// Leader election record annotation key defined.
							resourcelock.LeaderElectionRecordAnnotationKey: "this is okay",
						},
					},
					// Endpoint with zero subsets.
				},
				&v1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{
						// This won't contribute to any failures.
						Name:        "non-existent-service",
						Namespace:   "test",
						Annotations: map[string]string{},
					},
					// Endpoint with zero subsets.
				},
				&v1.Endpoints{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "Service1",
						Namespace:   "test",
						Annotations: map[string]string{},
					},
					// Endpoint with zero subsets.
				},
				&v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Service1",
						Namespace: "test",
					},
					Spec: v1.ServiceSpec{
						Selector: map[string]string{
							"app1": "test-app1",
							"app2": "test-app2",
						},
					},
				},
				&v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						// This service won't be discovered.
						Name:      "Service2",
						Namespace: "default",
					},
					Spec: v1.ServiceSpec{
						Selector: map[string]string{
							"app1": "test-app1",
							"app2": "test-app2",
						},
					},
				},
				&v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "Service3",
						Namespace: "test",
					},
					Spec: v1.ServiceSpec{
						// No Spec Selector
					},
				},
			),
		},
		Context:   context.Background(),
		Namespace: "test",
	}

	sAnalyzer := ServiceAnalyzer{}
	results, err := sAnalyzer.Analyze(config)
	require.NoError(t, err)

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	expectations := []struct {
		name          string
		failuresCount int
	}{
		{
			name:          "test/Endpoint1",
			failuresCount: 1,
		},
		{
			name:          "test/Service1",
			failuresCount: 2,
		},
	}

	require.Equal(t, len(expectations), len(results))

	for i, result := range results {
		require.Equal(t, expectations[i].name, result.Name)
		require.Equal(t, expectations[i].failuresCount, len(result.Error))
	}
}
