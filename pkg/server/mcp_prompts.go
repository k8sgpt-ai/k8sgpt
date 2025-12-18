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
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// getTroubleshootPodPrompt returns a prompt for pod troubleshooting
func (s *K8sGptMCPServer) getTroubleshootPodPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	podName := ""
	namespace := ""
	if request.Params.Arguments != nil {
		podName = request.Params.Arguments["podName"]
		namespace = request.Params.Arguments["namespace"]
	}

	promptText := fmt.Sprintf(`You are troubleshooting a Kubernetes pod issue.

Pod: %s
Namespace: %s

Troubleshooting steps:
1. Use 'get-resource' tool to get pod details and check status, conditions, and events
2. Use 'list-events' tool with the pod name to see recent events
3. Use 'get-logs' tool to check container logs for errors
4. Check if the pod has multiple containers and inspect each
5. If the pod is in CrashLoopBackOff, use 'get-logs' with previous=true
6. Use 'analyze' tool with filters=['Pod'] to get AI-powered analysis
7. Check related resources like ConfigMaps, Secrets, and PVCs

Common issues to check:
- Image pull errors (check imagePullSecrets)
- Resource limits (CPU/memory)
- Liveness/readiness probe failures
- Volume mount issues
- Environment variable problems
- Network connectivity issues`, podName, namespace)

	return &mcp.GetPromptResult{
		Description: "Pod troubleshooting guide",
		Messages: []mcp.PromptMessage{
			{
				Role: "user",
				Content: mcp.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}, nil
}

// getTroubleshootDeploymentPrompt returns a prompt for deployment troubleshooting
func (s *K8sGptMCPServer) getTroubleshootDeploymentPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	deploymentName := ""
	namespace := ""
	if request.Params.Arguments != nil {
		deploymentName = request.Params.Arguments["deploymentName"]
		namespace = request.Params.Arguments["namespace"]
	}

	promptText := fmt.Sprintf(`You are troubleshooting a Kubernetes deployment issue.

Deployment: %s
Namespace: %s

Troubleshooting steps:
1. Use 'get-resource' tool to get deployment details and check replica status
2. Use 'list-resources' with resourceType='replicasets' to check ReplicaSets
3. Use 'list-resources' with resourceType='pods' and labelSelector to find deployment pods
4. Use 'list-events' tool to see deployment-related events
5. Use 'analyze' tool with filters=['Deployment','Pod'] for comprehensive analysis
6. Check pod status and logs for individual pod issues
7. Verify image availability and pull secrets
8. Check resource quotas and limits

Common deployment issues:
- Insufficient resources in the cluster
- Image pull failures
- Invalid configuration (ConfigMaps/Secrets)
- Failed rolling updates
- Readiness probe failures preventing rollout
- PVC binding issues
- Node selector/affinity constraints`, deploymentName, namespace)

	return &mcp.GetPromptResult{
		Description: "Deployment troubleshooting guide",
		Messages: []mcp.PromptMessage{
			{
				Role: "user",
				Content: mcp.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}, nil
}

// getTroubleshootClusterPrompt returns a prompt for general cluster troubleshooting
func (s *K8sGptMCPServer) getTroubleshootClusterPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	promptText := `You are performing a general Kubernetes cluster health check and troubleshooting.

Recommended troubleshooting workflow:

1. CLUSTER OVERVIEW:
   - Use 'cluster-info' to get cluster version
   - Use 'list-namespaces' to see all namespaces
   - Use 'list-resources' with resourceType='nodes' to check node health

2. RESOURCE ANALYSIS:
   - Use 'analyze' tool with explain=true for comprehensive AI-powered analysis
   - Start with core resources: filters=['Pod','Deployment','Service']
   - Add more filters as needed: ['Node','PersistentVolumeClaim','Job','CronJob']

3. EVENT INSPECTION:
   - Use 'list-events' to see recent cluster events
   - Filter by namespace for focused troubleshooting
   - Look for Warning and Error events

4. SPECIFIC RESOURCE INVESTIGATION:
   - Use 'list-resources' to find problematic resources
   - Use 'get-resource' for detailed inspection
   - Use 'get-logs' to examine container logs

5. CONFIGURATION CHECK:
   - Use 'list-filters' to see available analyzers
   - Use 'list-integrations' to check integrations (Prometheus, AWS, etc.)
   - Use 'config' tool to modify settings if needed

Common cluster-wide issues:
- Node pressure (CPU, memory, disk)
- Network policies blocking traffic
- Storage provisioning problems
- RBAC permission issues
- Certificate expiration
- Control plane component failures
- Resource quota exhaustion
- DNS resolution problems

Use the available tools systematically to narrow down the issue.`

	return &mcp.GetPromptResult{
		Description: "General cluster troubleshooting guide",
		Messages: []mcp.PromptMessage{
			{
				Role: "user",
				Content: mcp.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}, nil
}
