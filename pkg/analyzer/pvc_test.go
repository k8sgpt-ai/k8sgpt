/*
Copyright 2024 The K8sGPT Authors.
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
	"time"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPersistentVolumeClaimAnalyzer(t *testing.T) {
	tests := []struct {
		name         string
		config       common.Analyzer
		expectations []string
	}{
		{
			name: "PV1 and PVC5 report failures",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&appsv1.Event{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "Event1",
								Namespace: "default",
							},
							LastTimestamp: metav1.Time{
								Time: time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC),
							},
							Reason:  "ProvisioningFailed",
							Message: "PVC Event1 provisioning failed",
						},
						&appsv1.Event{
							ObjectMeta: metav1.ObjectMeta{
								// This event won't get selected.
								Name:      "Event2",
								Namespace: "test",
							},
						},
						&appsv1.Event{
							// This is the latest event.
							ObjectMeta: metav1.ObjectMeta{
								Name:      "Event3",
								Namespace: "default",
							},
							LastTimestamp: metav1.Time{
								Time: time.Date(2024, 4, 15, 10, 0, 0, 0, time.UTC),
							},
							Reason:  "ProvisioningFailed",
							Message: "PVC Event3 provisioning failed",
						},
						&appsv1.PersistentVolumeClaim{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "PVC1",
								Namespace: "default",
							},
							Status: appsv1.PersistentVolumeClaimStatus{
								Phase: appsv1.ClaimPending,
							},
						},
						&appsv1.PersistentVolumeClaim{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "PVC2",
								Namespace: "default",
							},
							Status: appsv1.PersistentVolumeClaimStatus{
								// Won't contribute to failures.
								Phase: appsv1.ClaimBound,
							},
						},
						&appsv1.PersistentVolumeClaim{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "PVC3",
								Namespace: "default",
							},
							Status: appsv1.PersistentVolumeClaimStatus{
								// Won't contribute to failures.
								Phase: appsv1.ClaimLost,
							},
						},
						&appsv1.PersistentVolumeClaim{
							// PVCs in namespace other than "default" won't be discovered.
							ObjectMeta: metav1.ObjectMeta{
								Name:      "PVC4",
								Namespace: "test",
							},
							Status: appsv1.PersistentVolumeClaimStatus{
								Phase: appsv1.ClaimLost,
							},
						},
						&appsv1.PersistentVolumeClaim{
							// PVCs in namespace other than "default" won't be discovered.
							ObjectMeta: metav1.ObjectMeta{
								Name:      "PVC5",
								Namespace: "default",
							},
							Status: appsv1.PersistentVolumeClaimStatus{
								Phase: appsv1.ClaimPending,
							},
						},
					),
				},
				Context:   context.Background(),
				Namespace: "default",
			},
			expectations: []string{
				"default/PVC1",
				"default/PVC5",
			},
		},
		{
			name: "no event",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&appsv1.PersistentVolumeClaim{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "PVC1",
								Namespace: "default",
							},
							Status: appsv1.PersistentVolumeClaimStatus{
								Phase: appsv1.ClaimPending,
							},
						},
					),
				},
				Context:   context.Background(),
				Namespace: "default",
			},
		},
		{
			name: "event other than provision failure",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&appsv1.Event{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "Event1",
								Namespace: "default",
							},
							// Any reason other than ProvisioningFailed won't result in failure.
							Reason: "UnknownReason",
						},
						&appsv1.PersistentVolumeClaim{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "PVC1",
								Namespace: "default",
							},
							Status: appsv1.PersistentVolumeClaimStatus{
								Phase: appsv1.ClaimPending,
							},
						},
					),
				},
				Context:   context.Background(),
				Namespace: "default",
			},
		},
		{
			name: "event without error message",
			config: common.Analyzer{
				Client: &kubernetes.Client{
					Client: fake.NewSimpleClientset(
						&appsv1.Event{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "Event1",
								Namespace: "default",
							},
							// Event without any error message won't result in failure.
							Reason: "ProvisioningFailed",
						},
						&appsv1.PersistentVolumeClaim{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "PVC1",
								Namespace: "default",
							},
							Status: appsv1.PersistentVolumeClaimStatus{
								Phase: appsv1.ClaimPending,
							},
						},
					),
				},
				Context:   context.Background(),
				Namespace: "default",
			},
		},
	}

	pvcAnalyzer := PvcAnalyzer{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := pvcAnalyzer.Analyze(tt.config)
			require.NoError(t, err)

			if tt.expectations == nil {
				require.Equal(t, 0, len(results))
			} else {
				sort.Slice(results, func(i, j int) bool {
					return results[i].Name < results[j].Name
				})

				require.Equal(t, len(tt.expectations), len(results))

				for i, expectation := range tt.expectations {
					require.Equal(t, expectation, results[i].Name)
				}
			}
		})
	}
}
