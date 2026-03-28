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
	"fmt"
	"strings"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/spf13/viper"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type CRDAnalyzer struct{}

func (CRDAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	// Load CRD analyzer configuration
	var config common.CRDAnalyzerConfig
	if err := viper.UnmarshalKey("crd_analyzer", &config); err != nil {
		// If no config or error, disable the analyzer
		return nil, nil
	}

	if !config.Enabled {
		return nil, nil
	}

	// Create apiextensions client to discover CRDs
	apiExtClient, err := apiextensionsclientset.NewForConfig(a.Client.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to create apiextensions client: %w", err)
	}

	// List all CRDs in the cluster
	crdList, err := apiExtClient.ApiextensionsV1().CustomResourceDefinitions().List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list CRDs: %w", err)
	}

	var results []common.Result

	// Process each CRD
	for _, crd := range crdList.Items {
		// Check if CRD should be excluded
		if shouldExcludeCRD(crd.Name, config.Exclude) {
			continue
		}

		// Get the CRD configuration (if specified)
		crdConfig := getCRDConfig(crd.Name, config.Include)

		// Analyze resources for this CRD
		crdResults, err := analyzeCRDResources(a, crd, crdConfig)
		if err != nil {
			// Log error but continue with other CRDs
			continue
		}

		results = append(results, crdResults...)
	}

	return results, nil
}

// shouldExcludeCRD checks if a CRD should be excluded from analysis
func shouldExcludeCRD(crdName string, excludeList []common.CRDExcludeConfig) bool {
	for _, exclude := range excludeList {
		if exclude.Name == crdName {
			return true
		}
	}
	return false
}

// getCRDConfig returns the configuration for a specific CRD if it exists
func getCRDConfig(crdName string, includeList []common.CRDIncludeConfig) *common.CRDIncludeConfig {
	for _, include := range includeList {
		if include.Name == crdName {
			return &include
		}
	}
	return nil
}

// analyzeCRDResources analyzes all instances of a CRD
func analyzeCRDResources(a common.Analyzer, crd apiextensionsv1.CustomResourceDefinition, config *common.CRDIncludeConfig) ([]common.Result, error) {
	if a.Client.GetDynamicClient() == nil {
		return nil, fmt.Errorf("dynamic client is nil")
	}

	// Get the preferred version (typically the storage version)
	var version string
	for _, v := range crd.Spec.Versions {
		if v.Storage {
			version = v.Name
			break
		}
	}
	if version == "" && len(crd.Spec.Versions) > 0 {
		version = crd.Spec.Versions[0].Name
	}

	// Construct GVR
	gvr := schema.GroupVersionResource{
		Group:    crd.Spec.Group,
		Version:  version,
		Resource: crd.Spec.Names.Plural,
	}

	// List resources
	var list *unstructured.UnstructuredList
	var err error
	if crd.Spec.Scope == apiextensionsv1.NamespaceScoped {
		if a.Namespace != "" {
			list, err = a.Client.GetDynamicClient().Resource(gvr).Namespace(a.Namespace).List(a.Context, metav1.ListOptions{LabelSelector: a.LabelSelector})
		} else {
			list, err = a.Client.GetDynamicClient().Resource(gvr).Namespace(metav1.NamespaceAll).List(a.Context, metav1.ListOptions{LabelSelector: a.LabelSelector})
		}
	} else {
		// Cluster-scoped
		list, err = a.Client.GetDynamicClient().Resource(gvr).List(a.Context, metav1.ListOptions{LabelSelector: a.LabelSelector})
	}

	if err != nil {
		return nil, err
	}

	var results []common.Result

	// Analyze each resource instance
	for _, item := range list.Items {
		failures := analyzeResource(item, crd, config)
		if len(failures) > 0 {
			resourceName := item.GetName()
			if item.GetNamespace() != "" {
				resourceName = item.GetNamespace() + "/" + resourceName
			}

			results = append(results, common.Result{
				Kind:  crd.Spec.Names.Kind,
				Name:  resourceName,
				Error: failures,
			})
		}
	}

	return results, nil
}

// analyzeResource analyzes a single CR instance for issues
func analyzeResource(item unstructured.Unstructured, crd apiextensionsv1.CustomResourceDefinition, config *common.CRDIncludeConfig) []common.Failure {
	var failures []common.Failure

	// Check for deletion with finalizers (resource stuck in deletion)
	if item.GetDeletionTimestamp() != nil && len(item.GetFinalizers()) > 0 {
		failures = append(failures, common.Failure{
			Text: fmt.Sprintf("Resource is being deleted but has finalizers: %v", item.GetFinalizers()),
		})
	}

	// If custom config is provided, use it
	if config != nil {
		configFailures := analyzeWithConfig(item, config)
		failures = append(failures, configFailures...)
		return failures
	}

	// Otherwise, use generic health checks based on common patterns
	genericFailures := analyzeGenericHealth(item)
	failures = append(failures, genericFailures...)

	return failures
}

// analyzeWithConfig analyzes a resource using custom configuration
func analyzeWithConfig(item unstructured.Unstructured, config *common.CRDIncludeConfig) []common.Failure {
	var failures []common.Failure

	// Check ReadyCondition if specified
	if config.ReadyCondition != nil {
		conditions, found, err := unstructured.NestedSlice(item.Object, "status", "conditions")
		if !found || err != nil {
			failures = append(failures, common.Failure{
				Text: "Expected status.conditions not found",
			})
			return failures
		}

		ready := false
		var conditionMessages []string
		for _, cond := range conditions {
			condMap, ok := cond.(map[string]interface{})
			if !ok {
				continue
			}

			condType, _, _ := unstructured.NestedString(condMap, "type")
			status, _, _ := unstructured.NestedString(condMap, "status")
			message, _, _ := unstructured.NestedString(condMap, "message")

			if condType == config.ReadyCondition.Type {
				if status == config.ReadyCondition.ExpectedStatus {
					ready = true
				} else {
					conditionMessages = append(conditionMessages, fmt.Sprintf("%s=%s: %s", condType, status, message))
				}
			}
		}

		if !ready {
			msg := fmt.Sprintf("Ready condition not met: expected %s=%s", config.ReadyCondition.Type, config.ReadyCondition.ExpectedStatus)
			if len(conditionMessages) > 0 {
				msg += "; " + strings.Join(conditionMessages, "; ")
			}
			failures = append(failures, common.Failure{
				Text: msg,
			})
		}
	}

	// Check ExpectedValue if specified and StatusPath provided
	if config.ExpectedValue != "" && config.StatusPath != "" {
		pathParts := strings.Split(config.StatusPath, ".")
		// Remove leading dot if present
		if len(pathParts) > 0 && pathParts[0] == "" {
			pathParts = pathParts[1:]
		}

		actualValue, found, err := unstructured.NestedString(item.Object, pathParts...)
		if !found || err != nil {
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf("Expected field %s not found", config.StatusPath),
			})
		} else if actualValue != config.ExpectedValue {
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf("Field %s has value '%s', expected '%s'", config.StatusPath, actualValue, config.ExpectedValue),
			})
		}
	}

	return failures
}

// analyzeGenericHealth applies generic health checks based on common Kubernetes patterns
func analyzeGenericHealth(item unstructured.Unstructured) []common.Failure {
	var failures []common.Failure

	// Check for status.conditions (common pattern)
	conditions, found, err := unstructured.NestedSlice(item.Object, "status", "conditions")
	if found && err == nil && len(conditions) > 0 {
		for _, cond := range conditions {
			condMap, ok := cond.(map[string]interface{})
			if !ok {
				continue
			}

			condType, _, _ := unstructured.NestedString(condMap, "type")
			status, _, _ := unstructured.NestedString(condMap, "status")
			reason, _, _ := unstructured.NestedString(condMap, "reason")
			message, _, _ := unstructured.NestedString(condMap, "message")

			// Check for common failure patterns
			if condType == "Ready" && status != "True" {
				msg := fmt.Sprintf("Condition Ready is %s", status)
				if reason != "" {
					msg += fmt.Sprintf(" (reason: %s)", reason)
				}
				if message != "" {
					msg += fmt.Sprintf(": %s", message)
				}
				failures = append(failures, common.Failure{Text: msg})
			} else if strings.Contains(strings.ToLower(condType), "failed") && status == "True" {
				msg := fmt.Sprintf("Condition %s is True", condType)
				if message != "" {
					msg += fmt.Sprintf(": %s", message)
				}
				failures = append(failures, common.Failure{Text: msg})
			}
		}
	}

	// Check for status.phase (common pattern)
	phase, found, _ := unstructured.NestedString(item.Object, "status", "phase")
	if found && phase != "" {
		lowerPhase := strings.ToLower(phase)
		if lowerPhase == "failed" || lowerPhase == "error" {
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf("Resource phase is %s", phase),
			})
		}
	}

	// Check for status.health.status (ArgoCD pattern)
	healthStatus, found, _ := unstructured.NestedString(item.Object, "status", "health", "status")
	if found && healthStatus != "" {
		if healthStatus != "Healthy" && healthStatus != "Unknown" {
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf("Health status is %s", healthStatus),
			})
		}
	}

	// Check for status.state (common pattern)
	state, found, _ := unstructured.NestedString(item.Object, "status", "state")
	if found && state != "" {
		lowerState := strings.ToLower(state)
		if lowerState == "failed" || lowerState == "error" {
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf("Resource state is %s", state),
			})
		}
	}

	return failures
}
