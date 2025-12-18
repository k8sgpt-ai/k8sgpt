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

// handleListResources lists Kubernetes resources of a specific type
func (s *K8sGptMCPServer) handleListResources(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var req struct {
		ResourceType  string `json:"resourceType"`
		Namespace     string `json:"namespace,omitempty"`
		LabelSelector string `json:"labelSelector,omitempty"`
	}
	if err := request.BindArguments(&req); err != nil {
		return mcp.NewToolResultErrorf("Failed to parse request arguments: %v", err), nil
	}

	client, err := kubernetes.NewClient("", "")
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to create Kubernetes client: %v", err), nil
	}

	listOptions := metav1.ListOptions{}
	if req.LabelSelector != "" {
		listOptions.LabelSelector = req.LabelSelector
	}

	var result string
	resourceType := strings.ToLower(req.ResourceType)

	switch resourceType {
	case "pod", "pods":
		pods, err := client.Client.CoreV1().Pods(req.Namespace).List(ctx, listOptions)
		if err != nil {
			return mcp.NewToolResultErrorf("Failed to list pods: %v", err), nil
		}
		data, _ := json.MarshalIndent(pods.Items, "", "  ")
		result = string(data)

	case "deployment", "deployments":
		deps, err := client.Client.AppsV1().Deployments(req.Namespace).List(ctx, listOptions)
		if err != nil {
			return mcp.NewToolResultErrorf("Failed to list deployments: %v", err), nil
		}
		data, _ := json.MarshalIndent(deps.Items, "", "  ")
		result = string(data)

	case "service", "services", "svc":
		svcs, err := client.Client.CoreV1().Services(req.Namespace).List(ctx, listOptions)
		if err != nil {
			return mcp.NewToolResultErrorf("Failed to list services: %v", err), nil
		}
		data, _ := json.MarshalIndent(svcs.Items, "", "  ")
		result = string(data)

	case "node", "nodes":
		nodes, err := client.Client.CoreV1().Nodes().List(ctx, listOptions)
		if err != nil {
			return mcp.NewToolResultErrorf("Failed to list nodes: %v", err), nil
		}
		data, _ := json.MarshalIndent(nodes.Items, "", "  ")
		result = string(data)

	case "job", "jobs":
		jobs, err := client.Client.BatchV1().Jobs(req.Namespace).List(ctx, listOptions)
		if err != nil {
			return mcp.NewToolResultErrorf("Failed to list jobs: %v", err), nil
		}
		data, _ := json.MarshalIndent(jobs.Items, "", "  ")
		result = string(data)

	case "cronjob", "cronjobs":
		cronjobs, err := client.Client.BatchV1().CronJobs(req.Namespace).List(ctx, listOptions)
		if err != nil {
			return mcp.NewToolResultErrorf("Failed to list cronjobs: %v", err), nil
		}
		data, _ := json.MarshalIndent(cronjobs.Items, "", "  ")
		result = string(data)

	case "statefulset", "statefulsets", "sts":
		sts, err := client.Client.AppsV1().StatefulSets(req.Namespace).List(ctx, listOptions)
		if err != nil {
			return mcp.NewToolResultErrorf("Failed to list statefulsets: %v", err), nil
		}
		data, _ := json.MarshalIndent(sts.Items, "", "  ")
		result = string(data)

	case "daemonset", "daemonsets", "ds":
		ds, err := client.Client.AppsV1().DaemonSets(req.Namespace).List(ctx, listOptions)
		if err != nil {
			return mcp.NewToolResultErrorf("Failed to list daemonsets: %v", err), nil
		}
		data, _ := json.MarshalIndent(ds.Items, "", "  ")
		result = string(data)

	case "replicaset", "replicasets", "rs":
		rs, err := client.Client.AppsV1().ReplicaSets(req.Namespace).List(ctx, listOptions)
		if err != nil {
			return mcp.NewToolResultErrorf("Failed to list replicasets: %v", err), nil
		}
		data, _ := json.MarshalIndent(rs.Items, "", "  ")
		result = string(data)

	case "configmap", "configmaps", "cm":
		cms, err := client.Client.CoreV1().ConfigMaps(req.Namespace).List(ctx, listOptions)
		if err != nil {
			return mcp.NewToolResultErrorf("Failed to list configmaps: %v", err), nil
		}
		data, _ := json.MarshalIndent(cms.Items, "", "  ")
		result = string(data)

	case "secret", "secrets":
		secrets, err := client.Client.CoreV1().Secrets(req.Namespace).List(ctx, listOptions)
		if err != nil {
			return mcp.NewToolResultErrorf("Failed to list secrets: %v", err), nil
		}
		data, _ := json.MarshalIndent(secrets.Items, "", "  ")
		result = string(data)

	case "ingress", "ingresses", "ing":
		ingresses, err := client.Client.NetworkingV1().Ingresses(req.Namespace).List(ctx, listOptions)
		if err != nil {
			return mcp.NewToolResultErrorf("Failed to list ingresses: %v", err), nil
		}
		data, _ := json.MarshalIndent(ingresses.Items, "", "  ")
		result = string(data)

	case "persistentvolumeclaim", "persistentvolumeclaims", "pvc":
		pvcs, err := client.Client.CoreV1().PersistentVolumeClaims(req.Namespace).List(ctx, listOptions)
		if err != nil {
			return mcp.NewToolResultErrorf("Failed to list PVCs: %v", err), nil
		}
		data, _ := json.MarshalIndent(pvcs.Items, "", "  ")
		result = string(data)

	case "persistentvolume", "persistentvolumes", "pv":
		pvs, err := client.Client.CoreV1().PersistentVolumes().List(ctx, listOptions)
		if err != nil {
			return mcp.NewToolResultErrorf("Failed to list PVs: %v", err), nil
		}
		data, _ := json.MarshalIndent(pvs.Items, "", "  ")
		result = string(data)

	default:
		return mcp.NewToolResultErrorf("Unsupported resource type: %s. Supported types: pods, deployments, services, nodes, jobs, cronjobs, statefulsets, daemonsets, replicasets, configmaps, secrets, ingresses, pvc, pv", resourceType), nil
	}

	return mcp.NewToolResultText(result), nil
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

	client, err := kubernetes.NewClient("", "")
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to create Kubernetes client: %v", err), nil
	}

	var result string
	resourceType := strings.ToLower(req.ResourceType)

	switch resourceType {
	case "pod", "pods":
		pod, err := client.Client.CoreV1().Pods(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
		if err != nil {
			return mcp.NewToolResultErrorf("Failed to get pod: %v", err), nil
		}
		data, _ := json.MarshalIndent(pod, "", "  ")
		result = string(data)

	case "deployment", "deployments":
		dep, err := client.Client.AppsV1().Deployments(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
		if err != nil {
			return mcp.NewToolResultErrorf("Failed to get deployment: %v", err), nil
		}
		data, _ := json.MarshalIndent(dep, "", "  ")
		result = string(data)

	case "service", "services", "svc":
		svc, err := client.Client.CoreV1().Services(req.Namespace).Get(ctx, req.Name, metav1.GetOptions{})
		if err != nil {
			return mcp.NewToolResultErrorf("Failed to get service: %v", err), nil
		}
		data, _ := json.MarshalIndent(svc, "", "  ")
		result = string(data)

	case "node", "nodes":
		node, err := client.Client.CoreV1().Nodes().Get(ctx, req.Name, metav1.GetOptions{})
		if err != nil {
			return mcp.NewToolResultErrorf("Failed to get node: %v", err), nil
		}
		data, _ := json.MarshalIndent(node, "", "  ")
		result = string(data)

	default:
		return mcp.NewToolResultErrorf("Unsupported resource type: %s", resourceType), nil
	}

	return mcp.NewToolResultText(result), nil
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

	data, _ := json.MarshalIndent(namespaces.Items, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleListEvents lists Kubernetes events
func (s *K8sGptMCPServer) handleListEvents(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var req struct {
		Namespace           string `json:"namespace,omitempty"`
		InvolvedObjectName  string `json:"involvedObjectName,omitempty"`
		InvolvedObjectKind  string `json:"involvedObjectKind,omitempty"`
		Limit               int64  `json:"limit,omitempty"`
	}
	if err := request.BindArguments(&req); err != nil {
		return mcp.NewToolResultErrorf("Failed to parse request arguments: %v", err), nil
	}

	if req.Limit == 0 {
		req.Limit = 100
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

	data, _ := json.MarshalIndent(filteredEvents, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
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
	defer logStream.Close()

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

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleAddFilters adds filters to enable specific analyzers
func (s *K8sGptMCPServer) handleAddFilters(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var req struct {
		Filters []string `json:"filters"`
	}
	if err := request.BindArguments(&req); err != nil {
		return mcp.NewToolResultErrorf("Failed to parse request arguments: %v", err), nil
	}

	activeFilters := viper.GetStringSlice("active_filters")
	for _, filter := range req.Filters {
		if !contains(activeFilters, filter) {
			activeFilters = append(activeFilters, filter)
		}
	}

	viper.Set("active_filters", activeFilters)
	if err := viper.WriteConfig(); err != nil {
		return mcp.NewToolResultErrorf("Failed to save configuration: %v", err), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully added filters: %v", req.Filters)), nil
}

// handleRemoveFilters removes filters to disable specific analyzers
func (s *K8sGptMCPServer) handleRemoveFilters(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var req struct {
		Filters []string `json:"filters"`
	}
	if err := request.BindArguments(&req); err != nil {
		return mcp.NewToolResultErrorf("Failed to parse request arguments: %v", err), nil
	}

	activeFilters := viper.GetStringSlice("active_filters")
	newFilters := []string{}
	for _, filter := range activeFilters {
		if !contains(req.Filters, filter) {
			newFilters = append(newFilters, filter)
		}
	}

	viper.Set("active_filters", newFilters)
	if err := viper.WriteConfig(); err != nil {
		return mcp.NewToolResultErrorf("Failed to save configuration: %v", err), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully removed filters: %v", req.Filters)), nil
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

	data, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
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
