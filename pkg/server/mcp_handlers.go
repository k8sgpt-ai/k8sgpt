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

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/integration"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DefaultListLimit is the default maximum number of resources to return
	DefaultListLimit = 100
	// MaxListLimit is the maximum allowed limit for list operations
	MaxListLimit = 1000
)

// resourceLister defines a function that lists Kubernetes resources
type resourceLister func(ctx context.Context, client *kubernetes.Client, namespace string, opts metav1.ListOptions) (interface{}, error)

// resourceGetter defines a function that gets a single Kubernetes resource
type resourceGetter func(ctx context.Context, client *kubernetes.Client, namespace, name string) (interface{}, error)

// resourceRegistry maps resource types to their list and get functions
var resourceRegistry = map[string]struct {
	list resourceLister
	get  resourceGetter
}{
	"pod": {
		list: func(ctx context.Context, client *kubernetes.Client, namespace string, opts metav1.ListOptions) (interface{}, error) {
			return client.Client.CoreV1().Pods(namespace).List(ctx, opts)
		},
		get: func(ctx context.Context, client *kubernetes.Client, namespace, name string) (interface{}, error) {
			return client.Client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		},
	},
	"deployment": {
		list: func(ctx context.Context, client *kubernetes.Client, namespace string, opts metav1.ListOptions) (interface{}, error) {
			return client.Client.AppsV1().Deployments(namespace).List(ctx, opts)
		},
		get: func(ctx context.Context, client *kubernetes.Client, namespace, name string) (interface{}, error) {
			return client.Client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		},
	},
	"service": {
		list: func(ctx context.Context, client *kubernetes.Client, namespace string, opts metav1.ListOptions) (interface{}, error) {
			return client.Client.CoreV1().Services(namespace).List(ctx, opts)
		},
		get: func(ctx context.Context, client *kubernetes.Client, namespace, name string) (interface{}, error) {
			return client.Client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
		},
	},
	"node": {
		list: func(ctx context.Context, client *kubernetes.Client, namespace string, opts metav1.ListOptions) (interface{}, error) {
			return client.Client.CoreV1().Nodes().List(ctx, opts)
		},
		get: func(ctx context.Context, client *kubernetes.Client, namespace, name string) (interface{}, error) {
			return client.Client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
		},
	},
	"job": {
		list: func(ctx context.Context, client *kubernetes.Client, namespace string, opts metav1.ListOptions) (interface{}, error) {
			return client.Client.BatchV1().Jobs(namespace).List(ctx, opts)
		},
		get: func(ctx context.Context, client *kubernetes.Client, namespace, name string) (interface{}, error) {
			return client.Client.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
		},
	},
	"cronjob": {
		list: func(ctx context.Context, client *kubernetes.Client, namespace string, opts metav1.ListOptions) (interface{}, error) {
			return client.Client.BatchV1().CronJobs(namespace).List(ctx, opts)
		},
		get: func(ctx context.Context, client *kubernetes.Client, namespace, name string) (interface{}, error) {
			return client.Client.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
		},
	},
	"statefulset": {
		list: func(ctx context.Context, client *kubernetes.Client, namespace string, opts metav1.ListOptions) (interface{}, error) {
			return client.Client.AppsV1().StatefulSets(namespace).List(ctx, opts)
		},
		get: func(ctx context.Context, client *kubernetes.Client, namespace, name string) (interface{}, error) {
			return client.Client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		},
	},
	"daemonset": {
		list: func(ctx context.Context, client *kubernetes.Client, namespace string, opts metav1.ListOptions) (interface{}, error) {
			return client.Client.AppsV1().DaemonSets(namespace).List(ctx, opts)
		},
		get: func(ctx context.Context, client *kubernetes.Client, namespace, name string) (interface{}, error) {
			return client.Client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
		},
	},
	"replicaset": {
		list: func(ctx context.Context, client *kubernetes.Client, namespace string, opts metav1.ListOptions) (interface{}, error) {
			return client.Client.AppsV1().ReplicaSets(namespace).List(ctx, opts)
		},
		get: func(ctx context.Context, client *kubernetes.Client, namespace, name string) (interface{}, error) {
			return client.Client.AppsV1().ReplicaSets(namespace).Get(ctx, name, metav1.GetOptions{})
		},
	},
	"configmap": {
		list: func(ctx context.Context, client *kubernetes.Client, namespace string, opts metav1.ListOptions) (interface{}, error) {
			return client.Client.CoreV1().ConfigMaps(namespace).List(ctx, opts)
		},
		get: func(ctx context.Context, client *kubernetes.Client, namespace, name string) (interface{}, error) {
			return client.Client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
		},
	},
	"secret": {
		list: func(ctx context.Context, client *kubernetes.Client, namespace string, opts metav1.ListOptions) (interface{}, error) {
			return client.Client.CoreV1().Secrets(namespace).List(ctx, opts)
		},
		get: func(ctx context.Context, client *kubernetes.Client, namespace, name string) (interface{}, error) {
			return client.Client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
		},
	},
	"ingress": {
		list: func(ctx context.Context, client *kubernetes.Client, namespace string, opts metav1.ListOptions) (interface{}, error) {
			return client.Client.NetworkingV1().Ingresses(namespace).List(ctx, opts)
		},
		get: func(ctx context.Context, client *kubernetes.Client, namespace, name string) (interface{}, error) {
			return client.Client.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
		},
	},
	"persistentvolumeclaim": {
		list: func(ctx context.Context, client *kubernetes.Client, namespace string, opts metav1.ListOptions) (interface{}, error) {
			return client.Client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, opts)
		},
		get: func(ctx context.Context, client *kubernetes.Client, namespace, name string) (interface{}, error) {
			return client.Client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
		},
	},
	"persistentvolume": {
		list: func(ctx context.Context, client *kubernetes.Client, namespace string, opts metav1.ListOptions) (interface{}, error) {
			return client.Client.CoreV1().PersistentVolumes().List(ctx, opts)
		},
		get: func(ctx context.Context, client *kubernetes.Client, namespace, name string) (interface{}, error) {
			return client.Client.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
		},
	},
}

// Resource type aliases for convenience
var resourceTypeAliases = map[string]string{
	"pods":                      "pod",
	"deployments":               "deployment",
	"services":                  "service",
	"svc":                       "service",
	"nodes":                     "node",
	"jobs":                      "job",
	"cronjobs":                  "cronjob",
	"statefulsets":              "statefulset",
	"sts":                       "statefulset",
	"daemonsets":                "daemonset",
	"ds":                        "daemonset",
	"replicasets":               "replicaset",
	"rs":                        "replicaset",
	"configmaps":                "configmap",
	"cm":                        "configmap",
	"secrets":                   "secret",
	"ingresses":                 "ingress",
	"ing":                       "ingress",
	"persistentvolumeclaims":    "persistentvolumeclaim",
	"pvc":                       "persistentvolumeclaim",
	"persistentvolumes":         "persistentvolume",
	"pv":                        "persistentvolume",
}

// normalizeResourceType converts resource type variants to canonical form
func normalizeResourceType(resourceType string) (string, error) {
	normalized := strings.ToLower(resourceType)
	
	// Check if it's an alias
	if canonical, ok := resourceTypeAliases[normalized]; ok {
		normalized = canonical
	}
	
	// Check if it's a known resource type
	if _, ok := resourceRegistry[normalized]; !ok {
		return "", fmt.Errorf("unsupported resource type: %s", resourceType)
	}
	
	return normalized, nil
}

// marshalJSON marshals data to JSON with proper error handling
func marshalJSON(data interface{}) (string, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(jsonData), nil
}

// handleListResources lists Kubernetes resources of a specific type
func (s *K8sGptMCPServer) handleListResources(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var req struct {
		ResourceType  string `json:"resourceType"`
		Namespace     string `json:"namespace,omitempty"`
		LabelSelector string `json:"labelSelector,omitempty"`
		Limit         int64  `json:"limit,omitempty"`
	}
	if err := request.BindArguments(&req); err != nil {
		return mcp.NewToolResultErrorf("Failed to parse request arguments: %v", err), nil
	}

	if req.ResourceType == "" {
		return mcp.NewToolResultErrorf("resourceType is required"), nil
	}

	// Normalize and validate resource type
	resourceType, err := normalizeResourceType(req.ResourceType)
	if err != nil {
		supportedTypes := make([]string, 0, len(resourceRegistry))
		for key := range resourceRegistry {
			supportedTypes = append(supportedTypes, key)
		}
		return mcp.NewToolResultErrorf("%v. Supported types: %v", err, supportedTypes), nil
	}

	// Set default and validate limit
	if req.Limit == 0 {
		req.Limit = DefaultListLimit
	} else if req.Limit > MaxListLimit {
		req.Limit = MaxListLimit
	}

	client, err := kubernetes.NewClient("", "")
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to create Kubernetes client: %v", err), nil
	}

	listOptions := metav1.ListOptions{
		LabelSelector: req.LabelSelector,
		Limit:         req.Limit,
	}

	// Get the list function from registry
	listFunc := resourceRegistry[resourceType].list
	result, err := listFunc(ctx, client, req.Namespace, listOptions)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to list %s: %v", resourceType, err), nil
	}

	// Extract items from the result (all list types have an Items field)
	resultJSON, err := marshalJSON(result)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to serialize result: %v", err), nil
	}

	return mcp.NewToolResultText(resultJSON), nil
}

// handleGetResource gets detailed information about a specific resource
func (s *K8sGptMCPServer) handleGetResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var req struct {
		ResourceType string `json:"resourceType"`
		Name         string `json:"name"`
		Namespace    string `json:"namespace,omitempty"`
	}
	if err := request.BindArguments(&req); err != nil {
		return mcp.NewToolResultErrorf("Failed to parse request arguments: %v", err), nil
	}

	if req.ResourceType == "" {
		return mcp.NewToolResultErrorf("resourceType is required"), nil
	}
	if req.Name == "" {
		return mcp.NewToolResultErrorf("name is required"), nil
	}

	// Normalize and validate resource type
	resourceType, err := normalizeResourceType(req.ResourceType)
	if err != nil {
		return mcp.NewToolResultErrorf("%v", err), nil
	}

	client, err := kubernetes.NewClient("", "")
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to create Kubernetes client: %v", err), nil
	}

	// Get the get function from registry
	getFunc := resourceRegistry[resourceType].get
	result, err := getFunc(ctx, client, req.Namespace, req.Name)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to get %s '%s': %v", resourceType, req.Name, err), nil
	}

	resultJSON, err := marshalJSON(result)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to serialize result: %v", err), nil
	}

	return mcp.NewToolResultText(resultJSON), nil
}

// handleListNamespaces lists all namespaces in the cluster
func (s *K8sGptMCPServer) handleListNamespaces(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := kubernetes.NewClient("", "")
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to create Kubernetes client: %v", err), nil
	}

	namespaces, err := client.Client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to list namespaces: %v", err), nil
	}

	resultJSON, err := marshalJSON(namespaces.Items)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to serialize result: %v", err), nil
	}

	return mcp.NewToolResultText(resultJSON), nil
}

// handleListEvents lists Kubernetes events
func (s *K8sGptMCPServer) handleListEvents(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var req struct {
		Namespace          string `json:"namespace,omitempty"`
		InvolvedObjectName string `json:"involvedObjectName,omitempty"`
		InvolvedObjectKind string `json:"involvedObjectKind,omitempty"`
		Limit              int64  `json:"limit,omitempty"`
	}
	if err := request.BindArguments(&req); err != nil {
		return mcp.NewToolResultErrorf("Failed to parse request arguments: %v", err), nil
	}

	if req.Limit == 0 {
		req.Limit = DefaultListLimit
	} else if req.Limit > MaxListLimit {
		req.Limit = MaxListLimit
	}

	client, err := kubernetes.NewClient("", "")
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to create Kubernetes client: %v", err), nil
	}

	listOptions := metav1.ListOptions{
		Limit: req.Limit,
	}

	events, err := client.Client.CoreV1().Events(req.Namespace).List(ctx, listOptions)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to list events: %v", err), nil
	}

	// Filter events if needed
	filteredEvents := []corev1.Event{}
	for _, event := range events.Items {
		if req.InvolvedObjectName != "" && event.InvolvedObject.Name != req.InvolvedObjectName {
			continue
		}
		if req.InvolvedObjectKind != "" && event.InvolvedObject.Kind != req.InvolvedObjectKind {
			continue
		}
		filteredEvents = append(filteredEvents, event)
	}

	resultJSON, err := marshalJSON(filteredEvents)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to serialize result: %v", err), nil
	}

	return mcp.NewToolResultText(resultJSON), nil
}

// handleGetLogs retrieves logs from a pod container
func (s *K8sGptMCPServer) handleGetLogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var req struct {
		PodName      string `json:"podName"`
		Namespace    string `json:"namespace"`
		Container    string `json:"container,omitempty"`
		Previous     bool   `json:"previous,omitempty"`
		TailLines    int64  `json:"tailLines,omitempty"`
		SinceSeconds int64  `json:"sinceSeconds,omitempty"`
	}
	if err := request.BindArguments(&req); err != nil {
		return mcp.NewToolResultErrorf("Failed to parse request arguments: %v", err), nil
	}

	if req.PodName == "" {
		return mcp.NewToolResultErrorf("podName is required"), nil
	}
	if req.Namespace == "" {
		return mcp.NewToolResultErrorf("namespace is required"), nil
	}

	if req.TailLines == 0 {
		req.TailLines = 100
	}

	client, err := kubernetes.NewClient("", "")
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to create Kubernetes client: %v", err), nil
	}

	podLogOpts := &corev1.PodLogOptions{
		Container: req.Container,
		Previous:  req.Previous,
		TailLines: &req.TailLines,
	}

	if req.SinceSeconds > 0 {
		podLogOpts.SinceSeconds = &req.SinceSeconds
	}

	logRequest := client.Client.CoreV1().Pods(req.Namespace).GetLogs(req.PodName, podLogOpts)
	logStream, err := logRequest.Stream(ctx)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to get logs: %v", err), nil
	}
	defer func() {
		_ = logStream.Close()
	}()

	logs, err := io.ReadAll(logStream)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to read logs: %v", err), nil
	}

	return mcp.NewToolResultText(string(logs)), nil
}

// handleListFilters lists available and active filters
func (s *K8sGptMCPServer) handleListFilters(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	coreFilters, additionalFilters, integrationFilters := analyzer.ListFilters()
	active := viper.GetStringSlice("active_filters")

	result := map[string]interface{}{
		"coreFilters":        coreFilters,
		"additionalFilters":  additionalFilters,
		"integrationFilters": integrationFilters,
		"activeFilters":      active,
	}

	resultJSON, err := marshalJSON(result)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to serialize result: %v", err), nil
	}

	return mcp.NewToolResultText(resultJSON), nil
}

// handleAddFilters adds filters to enable specific analyzers
func (s *K8sGptMCPServer) handleAddFilters(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var req struct {
		Filters []string `json:"filters"`
	}
	if err := request.BindArguments(&req); err != nil {
		return mcp.NewToolResultErrorf("Failed to parse request arguments: %v", err), nil
	}

	if len(req.Filters) == 0 {
		return mcp.NewToolResultErrorf("filters array is required and cannot be empty"), nil
	}

	activeFilters := viper.GetStringSlice("active_filters")
	addedFilters := []string{}
	
	for _, filter := range req.Filters {
		if !contains(activeFilters, filter) {
			activeFilters = append(activeFilters, filter)
			addedFilters = append(addedFilters, filter)
		}
	}

	viper.Set("active_filters", activeFilters)
	if err := viper.WriteConfig(); err != nil {
		return mcp.NewToolResultErrorf("Failed to save configuration: %v", err), nil
	}

	if len(addedFilters) == 0 {
		return mcp.NewToolResultText("All specified filters were already active"), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully added filters: %v", addedFilters)), nil
}

// handleRemoveFilters removes filters to disable specific analyzers
func (s *K8sGptMCPServer) handleRemoveFilters(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var req struct {
		Filters []string `json:"filters"`
	}
	if err := request.BindArguments(&req); err != nil {
		return mcp.NewToolResultErrorf("Failed to parse request arguments: %v", err), nil
	}

	if len(req.Filters) == 0 {
		return mcp.NewToolResultErrorf("filters array is required and cannot be empty"), nil
	}

	activeFilters := viper.GetStringSlice("active_filters")
	newFilters := []string{}
	removedFilters := []string{}
	
	for _, filter := range activeFilters {
		if !contains(req.Filters, filter) {
			newFilters = append(newFilters, filter)
		} else {
			removedFilters = append(removedFilters, filter)
		}
	}

	viper.Set("active_filters", newFilters)
	if err := viper.WriteConfig(); err != nil {
		return mcp.NewToolResultErrorf("Failed to save configuration: %v", err), nil
	}

	if len(removedFilters) == 0 {
		return mcp.NewToolResultText("None of the specified filters were active"), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully removed filters: %v", removedFilters)), nil
}

// handleListIntegrations lists available integrations
func (s *K8sGptMCPServer) handleListIntegrations(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	integrationProvider := integration.NewIntegration()
	integrations := integrationProvider.List()

	result := []map[string]interface{}{}
	for _, integ := range integrations {
		active, _ := integrationProvider.IsActivate(integ)
		result = append(result, map[string]interface{}{
			"name":   integ,
			"active": active,
		})
	}

	resultJSON, err := marshalJSON(result)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to serialize result: %v", err), nil
	}

	return mcp.NewToolResultText(resultJSON), nil
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
