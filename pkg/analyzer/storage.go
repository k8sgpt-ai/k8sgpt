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
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StorageAnalyzer struct{}

func (StorageAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	kind := "Storage"

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	var results []common.Result

	// Analyze StorageClasses
	scResults, err := analyzeStorageClasses(a)
	if err != nil {
		return nil, err
	}
	results = append(results, scResults...)

	// Analyze PersistentVolumes
	pvResults, err := analyzePersistentVolumes(a)
	if err != nil {
		return nil, err
	}
	results = append(results, pvResults...)

	// Analyze PVCs with enhanced checks
	pvcResults, err := analyzePersistentVolumeClaims(a)
	if err != nil {
		return nil, err
	}
	results = append(results, pvcResults...)

	return results, nil
}

func analyzeStorageClasses(a common.Analyzer) ([]common.Result, error) {
	var results []common.Result

	scs, err := a.Client.GetClient().StorageV1().StorageClasses().List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, sc := range scs.Items {
		var failures []common.Failure

		// Check for deprecated storage classes
		if sc.Provisioner == "kubernetes.io/no-provisioner" {
			failures = append(failures, common.Failure{
				Text:      fmt.Sprintf("StorageClass %s uses deprecated provisioner 'kubernetes.io/no-provisioner'", sc.Name),
				Sensitive: []common.Sensitive{},
			})
		}

		// Check for default storage class
		if sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
			// Check if there are multiple default storage classes
			defaultCount := 0
			for _, otherSc := range scs.Items {
				if otherSc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
					defaultCount++
				}
			}
			if defaultCount > 1 {
				failures = append(failures, common.Failure{
					Text:      fmt.Sprintf("Multiple default StorageClasses found (%d), which can cause confusion", defaultCount),
					Sensitive: []common.Sensitive{},
				})
			}
		}

		if len(failures) > 0 {
			results = append(results, common.Result{
				Kind:  "Storage/StorageClass",
				Name:  sc.Name,
				Error: failures,
			})
			AnalyzerErrorsMetric.WithLabelValues("Storage/StorageClass", sc.Name, "").Set(float64(len(failures)))
		}
	}

	return results, nil
}

func analyzePersistentVolumes(a common.Analyzer) ([]common.Result, error) {
	var results []common.Result

	pvs, err := a.Client.GetClient().CoreV1().PersistentVolumes().List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pv := range pvs.Items {
		var failures []common.Failure

		// Check for released PVs
		if pv.Status.Phase == v1.VolumeReleased {
			failures = append(failures, common.Failure{
				Text:      fmt.Sprintf("PersistentVolume %s is in Released state and should be cleaned up", pv.Name),
				Sensitive: []common.Sensitive{},
			})
		}

		// Check for failed PVs
		if pv.Status.Phase == v1.VolumeFailed {
			failures = append(failures, common.Failure{
				Text:      fmt.Sprintf("PersistentVolume %s is in Failed state", pv.Name),
				Sensitive: []common.Sensitive{},
			})
		}

		// Check for small PVs (less than 1Gi)
		if capacity, ok := pv.Spec.Capacity[v1.ResourceStorage]; ok {
			if capacity.Cmp(resource.MustParse("1Gi")) < 0 {
				failures = append(failures, common.Failure{
					Text:      fmt.Sprintf("PersistentVolume %s has small capacity (%s)", pv.Name, capacity.String()),
					Sensitive: []common.Sensitive{},
				})
			}
		}

		if len(failures) > 0 {
			results = append(results, common.Result{
				Kind:  "Storage/PersistentVolume",
				Name:  pv.Name,
				Error: failures,
			})
			AnalyzerErrorsMetric.WithLabelValues("Storage/PersistentVolume", pv.Name, "").Set(float64(len(failures)))
		}
	}

	return results, nil
}

func analyzePersistentVolumeClaims(a common.Analyzer) ([]common.Result, error) {
	var results []common.Result

	pvcs, err := a.Client.GetClient().CoreV1().PersistentVolumeClaims(a.Namespace).List(a.Context, metav1.ListOptions{
		LabelSelector: a.LabelSelector,
	})
	if err != nil {
		return nil, err
	}

	for _, pvc := range pvcs.Items {
		var failures []common.Failure

		// Check for PVC state issues first (most critical)
		switch pvc.Status.Phase {
		case v1.ClaimPending:
			failures = append(failures, common.Failure{
				Text:      fmt.Sprintf("PersistentVolumeClaim %s is in Pending state", pvc.Name),
				Sensitive: []common.Sensitive{},
			})
		case v1.ClaimLost:
			failures = append(failures, common.Failure{
				Text:      fmt.Sprintf("PersistentVolumeClaim %s is in Lost state", pvc.Name),
				Sensitive: []common.Sensitive{},
			})
		default:
			// Only check other issues if PVC is not in a critical state
			if capacity, ok := pvc.Spec.Resources.Requests[v1.ResourceStorage]; ok {
				if capacity.Cmp(resource.MustParse("1Gi")) < 0 {
					failures = append(failures, common.Failure{
						Text:      fmt.Sprintf("PersistentVolumeClaim %s has small capacity (%s)", pvc.Name, capacity.String()),
						Sensitive: []common.Sensitive{},
					})
				}
			}

			// Check for missing storage class
			if pvc.Spec.StorageClassName == nil && pvc.Spec.VolumeName == "" {
				failures = append(failures, common.Failure{
					Text:      fmt.Sprintf("PersistentVolumeClaim %s has no StorageClass specified", pvc.Name),
					Sensitive: []common.Sensitive{},
				})
			}
		}

		// Only report the first failure found
		if len(failures) > 0 {
			results = append(results, common.Result{
				Kind:  "Storage/PersistentVolumeClaim",
				Name:  fmt.Sprintf("%s/%s", pvc.Namespace, pvc.Name),
				Error: failures[:1],
			})
			AnalyzerErrorsMetric.WithLabelValues("Storage/PersistentVolumeClaim", pvc.Name, pvc.Namespace).Set(1)
		}
	}

	return results, nil
}
