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
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestStorageAnalyzer(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		storageClasses []storagev1.StorageClass
		pvs            []v1.PersistentVolume
		pvcs           []v1.PersistentVolumeClaim
		expectedErrors int
	}{
		{
			name:      "Deprecated StorageClass",
			namespace: "default",
			storageClasses: []storagev1.StorageClass{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deprecated-sc",
					},
					Provisioner: "kubernetes.io/no-provisioner",
				},
			},
			expectedErrors: 1,
		},
		{
			name:      "Multiple Default StorageClasses",
			namespace: "default",
			storageClasses: []storagev1.StorageClass{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "default-sc1",
						Annotations: map[string]string{
							"storageclass.kubernetes.io/is-default-class": "true",
						},
					},
					Provisioner: "kubernetes.io/gce-pd",
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "default-sc2",
						Annotations: map[string]string{
							"storageclass.kubernetes.io/is-default-class": "true",
						},
					},
					Provisioner: "kubernetes.io/aws-ebs",
				},
			},
			expectedErrors: 2,
		},
		{
			name:      "Released PV",
			namespace: "default",
			pvs: []v1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "released-pv",
					},
					Status: v1.PersistentVolumeStatus{
						Phase: v1.VolumeReleased,
					},
				},
			},
			expectedErrors: 1,
		},
		{
			name:      "Failed PV",
			namespace: "default",
			pvs: []v1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "failed-pv",
					},
					Status: v1.PersistentVolumeStatus{
						Phase: v1.VolumeFailed,
					},
				},
			},
			expectedErrors: 1,
		},
		{
			name:      "Small PV",
			namespace: "default",
			pvs: []v1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "small-pv",
					},
					Spec: v1.PersistentVolumeSpec{
						Capacity: v1.ResourceList{
							v1.ResourceStorage: resource.MustParse("500Mi"),
						},
					},
				},
			},
			expectedErrors: 1,
		},
		{
			name:      "Pending PVC",
			namespace: "default",
			pvcs: []v1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pending-pvc",
						Namespace: "default",
					},
					Status: v1.PersistentVolumeClaimStatus{
						Phase: v1.ClaimPending,
					},
				},
			},
			expectedErrors: 1,
		},
		{
			name:      "Lost PVC",
			namespace: "default",
			pvcs: []v1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "lost-pvc",
						Namespace: "default",
					},
					Status: v1.PersistentVolumeClaimStatus{
						Phase: v1.ClaimLost,
					},
				},
			},
			expectedErrors: 1,
		},
		{
			name:      "Small PVC",
			namespace: "default",
			pvcs: []v1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "small-pvc",
						Namespace: "default",
					},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.VolumeResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse("500Mi"),
							},
						},
					},
				},
			},
			expectedErrors: 1,
		},
		{
			name:      "PVC without StorageClass",
			namespace: "default",
			pvcs: []v1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "no-sc-pvc",
						Namespace: "default",
					},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.VolumeResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client
			client := fake.NewSimpleClientset()

			// Create test resources
			for _, sc := range tt.storageClasses {
				_, err := client.StorageV1().StorageClasses().Create(context.TODO(), &sc, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("Failed to create StorageClass: %v", err)
				}
			}

			for _, pv := range tt.pvs {
				_, err := client.CoreV1().PersistentVolumes().Create(context.TODO(), &pv, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("Failed to create PV: %v", err)
				}
			}

			for _, pvc := range tt.pvcs {
				_, err := client.CoreV1().PersistentVolumeClaims(tt.namespace).Create(context.TODO(), &pvc, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("Failed to create PVC: %v", err)
				}
			}

			// Create analyzer
			analyzer := StorageAnalyzer{}

			// Create analyzer config
			config := common.Analyzer{
				Client: &kubernetes.Client{
					Client: client,
				},
				Context:   context.TODO(),
				Namespace: tt.namespace,
			}

			// Run analysis
			results, err := analyzer.Analyze(config)
			if err != nil {
				t.Fatalf("Failed to run analysis: %v", err)
			}

			// Count total errors
			totalErrors := 0
			for _, result := range results {
				totalErrors += len(result.Error)
			}

			// Check error count
			if totalErrors != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectedErrors, totalErrors)
			}
		})
	}
}
