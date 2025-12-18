# K8sGPT MCP Server Documentation

This document describes the Model Context Protocol (MCP) server implementation for k8sgpt, which enables AI agents and tools to interact with Kubernetes clusters through k8sgpt's powerful analysis capabilities.

## Overview

The K8sGPT MCP Server exposes k8sgpt functionality through the Model Context Protocol, allowing AI assistants like Claude, ChatGPT, and others to:
- Analyze Kubernetes resources for issues
- List and inspect cluster resources
- View logs and events
- Manage k8sgpt filters and integrations
- Access troubleshooting guides

## Features

### Tools (12 total)

#### Core Analysis
- **analyze**: Run k8sgpt analysis on Kubernetes resources with AI-powered explanations
- **cluster-info**: Get Kubernetes cluster version and information
- **config**: Configure k8sgpt settings (custom analyzers, cache)

#### Resource Management
- **list-resources**: List Kubernetes resources (pods, deployments, services, nodes, jobs, cronjobs, statefulsets, daemonsets, replicasets, configmaps, secrets, ingresses, PVCs, PVs)
- **get-resource**: Get detailed information about a specific resource
- **list-namespaces**: List all namespaces in the cluster

#### Debugging & Troubleshooting
- **list-events**: List Kubernetes events with filtering by object name/kind
- **get-logs**: Retrieve container logs from pods with options for previous logs, tail lines, and time filtering

#### Filter Management
- **list-filters**: List all available and active analyzers/filters
- **add-filters**: Enable specific analyzers
- **remove-filters**: Disable specific analyzers

#### Integration Management
- **list-integrations**: List available integrations (Prometheus, AWS, Keda, Kyverno)

### Resources (3 total)

- **cluster-info**: Quick access to cluster version information
- **namespaces**: List of all namespaces
- **active-filters**: Currently enabled analyzers

### Prompts (3 total)

Guided troubleshooting workflows:
- **troubleshoot-pod**: Step-by-step guide for diagnosing pod issues
- **troubleshoot-deployment**: Deployment troubleshooting workflow
- **troubleshoot-cluster**: General cluster health check and troubleshooting guide

## Usage

### Starting the MCP Server

#### As Part of k8sgpt serve Command

```bash
# Start with MCP enabled (stdio mode)
k8sgpt serve --enable-mcp

# Start with MCP in HTTP mode
k8sgpt serve --enable-mcp --mcp-http --mcp-port 8089
```

#### Standalone Mode

```bash
# Using the example binary
cd pkg/server/example
go run main.go --http --port 8089
```

### Configuration

The MCP server respects all k8sgpt configuration including:
- Active filters from `~/.config/k8sgpt/k8sgpt.yaml`
- AI backend configuration
- Custom analyzers
- Cache settings
- Kubernetes context

### Transport Modes

1. **Stdio Mode** (default): Communicates via standard input/output
   - Best for local AI assistants
   - Used by Claude Desktop, VS Code extensions, etc.

2. **HTTP Mode**: Runs as HTTP server with SSE streaming
   - Accessible over network
   - Supports multiple concurrent clients
   - Useful for web-based AI tools

## Tool Examples

### Analyze Resources

```json
{
  "name": "analyze",
  "arguments": {
    "namespace": "default",
    "filters": ["Pod", "Deployment"],
    "explain": true
  }
}
```

### List Pods

```json
{
  "name": "list-resources",
  "arguments": {
    "resourceType": "pods",
    "namespace": "kube-system",
    "labelSelector": "k8s-app=kube-dns"
  }
}
```

### Get Logs

```json
{
  "name": "get-logs",
  "arguments": {
    "podName": "my-app-pod",
    "namespace": "default",
    "container": "my-container",
    "tailLines": 50,
    "previous": false
  }
}
```

### List Events

```json
{
  "name": "list-events",
  "arguments": {
    "namespace": "default",
    "involvedObjectName": "my-deployment",
    "involvedObjectKind": "Deployment",
    "limit": 50
  }
}
```

### Manage Filters

```json
// List available filters
{
  "name": "list-filters",
  "arguments": {}
}

// Add filters
{
  "name": "add-filters",
  "arguments": {
    "filters": ["Pod", "Service", "Ingress"]
  }
}

// Remove filters
{
  "name": "remove-filters",
  "arguments": {
    "filters": ["Service"]
  }
}
```

## Integration with AI Assistants

### Claude Desktop

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "k8sgpt": {
      "command": "k8sgpt",
      "args": ["serve", "--enable-mcp"]
    }
  }
}
```

### VS Code

Use with MCP-compatible VS Code extensions by pointing to the k8sgpt server.

### Custom Implementations

The HTTP mode allows integration with any MCP-compatible client:

```bash
# Start HTTP server
k8sgpt serve --enable-mcp --mcp-http --mcp-port 8089

# Connect from your client
curl http://localhost:8089/sse
```

## Prompts Usage

Prompts provide contextual guidance for common troubleshooting scenarios:

### Pod Troubleshooting

```json
{
  "name": "troubleshoot-pod",
  "arguments": {
    "podName": "failing-pod",
    "namespace": "production"
  }
}
```

Returns a structured guide with:
- Steps to diagnose the issue
- Relevant tools to use
- Common problems to check
- Best practices

### Deployment Troubleshooting

```json
{
  "name": "troubleshoot-deployment",
  "arguments": {
    "deploymentName": "my-app",
    "namespace": "production"
  }
}
```

### General Cluster Troubleshooting

```json
{
  "name": "troubleshoot-cluster",
  "arguments": {}
}
```

Provides a comprehensive workflow for cluster-wide health checks.

## Resource Types Supported

The `list-resources` and `get-resource` tools support:

| Resource Type | Aliases | Scope |
|--------------|---------|-------|
| Pod | pods | Namespaced |
| Deployment | deployments | Namespaced |
| Service | services, svc | Namespaced |
| Node | nodes | Cluster |
| Job | jobs | Namespaced |
| CronJob | cronjobs | Namespaced |
| StatefulSet | statefulsets, sts | Namespaced |
| DaemonSet | daemonsets, ds | Namespaced |
| ReplicaSet | replicasets, rs | Namespaced |
| ConfigMap | configmaps, cm | Namespaced |
| Secret | secrets | Namespaced |
| Ingress | ingresses, ing | Namespaced |
| PersistentVolumeClaim | persistentvolumeclaims, pvc | Namespaced |
| PersistentVolume | persistentvolumes, pv | Cluster |

## Error Handling

All tools return structured error responses when:
- Kubernetes client fails to initialize
- Resources are not found
- Invalid parameters are provided
- Insufficient RBAC permissions

Error responses include descriptive messages to help diagnose issues.

## Performance Considerations

- **List operations**: Large clusters may have many resources; use label selectors to filter
- **Log retrieval**: Use `tailLines` and `sinceSeconds` to limit log volume
- **Event listing**: Default limit is 100 events; adjust as needed
- **Analysis**: Use specific filters to reduce analysis time and token usage

## Security Notes

1. **RBAC**: The MCP server uses the kubeconfig context and respects RBAC
2. **Secrets**: While the `list-resources` tool can list secrets, their data is returned as-is from the Kubernetes API
3. **Network**: HTTP mode exposes the server on a network port - use appropriate security measures
4. **AI Context**: Be mindful of sensitive data when using AI-powered analysis

## Troubleshooting

### MCP Server Won't Start

```bash
# Check if k8sgpt can connect to cluster
k8sgpt auth list

# Verify kubeconfig
kubectl cluster-info

# Check logs
k8sgpt serve --enable-mcp --verbose
```

### Tools Not Working

- Ensure active filters are set for analysis: `k8sgpt filters add Pod Deployment`
- Check RBAC permissions for resource types
- Verify namespace exists for namespaced resources

### HTTP Mode Issues

```bash
# Check if port is available
netstat -an | grep 8089

# Test connectivity
curl http://localhost:8089/health
```

## Development

### Adding New Tools

1. Define tool in `registerToolsAndResources()` in [mcp.go](mcp.go)
2. Implement handler function in [mcp_handlers.go](mcp_handlers.go)
3. Add request/response types if needed
4. Update documentation

### Adding New Resources

1. Define resource in `registerResources()` in [mcp.go](mcp.go)
2. Implement resource handler function
3. Update documentation

### Adding New Prompts

1. Define prompt in `registerPrompts()` in [mcp.go](mcp.go)
2. Implement prompt handler in [mcp_prompts.go](mcp_prompts.go)
3. Update documentation

## Architecture

```
┌─────────────────┐
│   AI Assistant  │ (Claude, ChatGPT, etc.)
└────────┬────────┘
         │ MCP Protocol
         │ (stdio or HTTP/SSE)
         ▼
┌─────────────────┐
│  MCP Server     │
│  (k8sgpt)       │
├─────────────────┤
│ • Tools         │
│ • Resources     │
│ • Prompts       │
└────────┬────────┘
         │ Kubernetes Client-Go
         ▼
┌─────────────────┐
│  K8s Cluster    │
└─────────────────┘
```

## Future Enhancements

Potential areas for expansion:
- [ ] Resource creation/update/delete tools
- [ ] Pod exec/port-forward capabilities
- [ ] Real-time resource watching
- [ ] Metrics querying (Prometheus integration)
- [ ] Cost analysis tools
- [ ] Security scanning tools
- [ ] Cluster backup/restore operations
- [ ] GitOps integration
- [ ] Multi-cluster support

## Contributing

Contributions are welcome! Please:
1. Follow the existing code structure
2. Add tests for new features
3. Update documentation
4. Ensure backward compatibility
5. Follow Go best practices

## License

Apache License 2.0 - See LICENSE file for details
