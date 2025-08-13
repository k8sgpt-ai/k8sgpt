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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigMapAnalyzer struct{}

func (ConfigMapAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	kind := "ConfigMap"

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	// Get all ConfigMaps in the namespace
	configMaps, err := a.Client.GetClient().CoreV1().ConfigMaps(a.Namespace).List(a.Context, metav1.ListOptions{
		LabelSelector: a.LabelSelector,
	})
	if err != nil {
		return nil, err
	}

	// Get all Pods to check ConfigMap usage
	pods, err := a.Client.GetClient().CoreV1().Pods(a.Namespace).List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var results []common.Result

	// Track which ConfigMaps are used
	usedConfigMaps := make(map[string]bool)
	configMapUsage := make(map[string][]string) // maps ConfigMap name to list of pods using it

	// Analyze ConfigMap usage in Pods
	for _, pod := range pods.Items {
		// Check volume mounts
		for _, volume := range pod.Spec.Volumes {
			if volume.ConfigMap != nil {
				usedConfigMaps[volume.ConfigMap.Name] = true
				configMapUsage[volume.ConfigMap.Name] = append(configMapUsage[volume.ConfigMap.Name], pod.Name)
			}
		}

		// Check environment variables
		for _, container := range pod.Spec.Containers {
			for _, env := range container.EnvFrom {
				if env.ConfigMapRef != nil {
					usedConfigMaps[env.ConfigMapRef.Name] = true
					configMapUsage[env.ConfigMapRef.Name] = append(configMapUsage[env.ConfigMapRef.Name], pod.Name)
				}
			}
			for _, env := range container.Env {
				if env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil {
					usedConfigMaps[env.ValueFrom.ConfigMapKeyRef.Name] = true
					configMapUsage[env.ValueFrom.ConfigMapKeyRef.Name] = append(configMapUsage[env.ValueFrom.ConfigMapKeyRef.Name], pod.Name)
				}
			}
		}
	}

	// Analyze each ConfigMap
	for _, cm := range configMaps.Items {
		var failures []common.Failure

		// Check for unused ConfigMaps
		if !usedConfigMaps[cm.Name] {
			failures = append(failures, common.Failure{
				Text:      fmt.Sprintf("ConfigMap %s is not used by any pods in the namespace", cm.Name),
				Sensitive: []common.Sensitive{},
			})
		}

		// Check for empty ConfigMaps
		if len(cm.Data) == 0 && len(cm.BinaryData) == 0 {
			failures = append(failures, common.Failure{
				Text:      fmt.Sprintf("ConfigMap %s is empty", cm.Name),
				Sensitive: []common.Sensitive{},
			})
		}

		// Check for large ConfigMaps (over 1MB)
		totalSize := 0
		for _, value := range cm.Data {
			totalSize += len(value)
		}
		for _, value := range cm.BinaryData {
			totalSize += len(value)
		}
		if totalSize > 1024*1024 { // 1MB
			failures = append(failures, common.Failure{
				Text:      fmt.Sprintf("ConfigMap %s is larger than 1MB (%d bytes)", cm.Name, totalSize),
				Sensitive: []common.Sensitive{},
			})
		}

		if len(failures) > 0 {
			results = append(results, common.Result{
				Kind:  kind,
				Name:  fmt.Sprintf("%s/%s", cm.Namespace, cm.Name),
				Error: failures,
			})
			AnalyzerErrorsMetric.WithLabelValues(kind, cm.Name, cm.Namespace).Set(float64(len(failures)))
		}
	}

	return results, nil
}
