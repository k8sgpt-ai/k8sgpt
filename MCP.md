# K8sGPT Model Context Protocol (MCP) Server

K8sGPT provides a Model Context Protocol (MCP) server that exposes Kubernetes cluster operations as standardized tools, resources, and prompts for AI assistants like Claude, ChatGPT, and other MCP-compatible clients.

## Table of Contents

- [What is MCP?](#what-is-mcp)
- [Quick Start](#quick-start)
- [Server Modes](#server-modes)
- [Available Tools](#available-tools)
- [Available Resources](#available-resources)
- [Available Prompts](#available-prompts)
- [Usage Examples](#usage-examples)
- [Integration with AI Assistants](#integration-with-ai-assistants)
- [HTTP API Reference](#http-api-reference)

## What is MCP?

The Model Context Protocol (MCP) is an open standard that enables AI assistants to securely connect to external data sources and tools. K8sGPT's MCP server exposes Kubernetes operations through this standardized interface, allowing AI assistants to:

- Analyze cluster health and issues
- Query Kubernetes resources
- Access pod logs and events
- Get troubleshooting guidance
- Manage analyzer filters

## Quick Start

### Start the MCP Server

**Stdio mode (for local AI assistants):**
```bash
k8sgpt serve --mcp
```

**HTTP mode (for network access):**
```bash
k8sgpt serve --mcp --mcp-http --mcp-port 8089
```

### Test with curl

```bash
curl -X POST http://localhost:8089/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/list"
  }'
```

## Server Modes

### Stdio Mode (Default)

Used by local AI assistants like Claude Desktop:

```bash
k8sgpt serve --mcp
```

Configure in your MCP client (e.g., Claude Desktop's `claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "k8sgpt": {
      "command": "k8sgpt",
      "args": ["serve", "--mcp"]
    }
  }
}
```

### HTTP Mode

Used for network access and webhooks:

```bash
k8sgpt serve --mcp --mcp-http --mcp-port 8089
```

The server runs in stateless mode, so no session management is required. Each request is independent.

## Available Tools

The MCP server exposes 12 tools for Kubernetes operations:

### Cluster Analysis

**analyze**
- Analyze Kubernetes resources for issues and problems
- Parameters:
  - `namespace` (optional): Namespace to analyze
  - `explain` (optional): Get AI explanations for issues
  - `filters` (optional): Comma-separated list of analyzers to use

**cluster-info**
- Get Kubernetes cluster information and version

### Resource Management

**list-resources**
- List Kubernetes resources of a specific type
- Parameters:
  - `resourceType` (required): Type of resource (pods, deployments, services, nodes, jobs, cronjobs, statefulsets, daemonsets, replicasets, configmaps, secrets, ingresses, pvcs, pvs)
  - `namespace` (optional): Namespace to query
  - `labelSelector` (optional): Label selector for filtering

**get-resource**
- Get detailed information about a specific Kubernetes resource
- Parameters:
  - `resourceType` (required): Type of resource
  - `name` (required): Resource name
  - `namespace` (optional): Namespace

**list-namespaces**
- List all namespaces in the cluster

### Debugging and Troubleshooting

**get-logs**
- Get logs from a pod container
- Parameters:
  - `podName` (required): Name of the pod
  - `namespace` (optional): Namespace
  - `container` (optional): Container name
  - `tail` (optional): Number of lines to show
  - `previous` (optional): Show logs from previous container instance
  - `sinceSeconds` (optional): Show logs from last N seconds

**list-events**
- List Kubernetes events for debugging
- Parameters:
  - `namespace` (optional): Namespace to query
  - `involvedObjectName` (optional): Filter by object name
  - `involvedObjectKind` (optional): Filter by object kind

### Analyzer Management

**list-filters**
- List all available and active analyzers/filters

**add-filters**
- Add filters to enable specific analyzers
- Parameters:
  - `filters` (required): Comma-separated list of analyzer names

**remove-filters**
- Remove filters to disable specific analyzers
- Parameters:
  - `filters` (required): Comma-separated list of analyzer names

### Integrations

**list-integrations**
- List available integrations (Prometheus, AWS, Keda, Kyverno, etc.)

### Configuration

**config**
- Configure K8sGPT settings including custom analyzers and cache

## Available Resources

Resources provide read-only access to cluster information:

**cluster-info**
- URI: `cluster-info`
- Get information about the Kubernetes cluster

**namespaces**
- URI: `namespaces`
- List all namespaces in the cluster

**active-filters**
- URI: `active-filters`
- Get currently active analyzers/filters

## Available Prompts

Prompts provide guided troubleshooting workflows:

**troubleshoot-pod**
- Interactive pod debugging workflow
- Arguments:
  - `podName` (required): Name of the pod to troubleshoot
  - `namespace` (required): Namespace of the pod

**troubleshoot-deployment**
- Interactive deployment debugging workflow
- Arguments:
  - `deploymentName` (required): Name of the deployment
  - `namespace` (required): Namespace of the deployment

**troubleshoot-cluster**
- General cluster troubleshooting workflow

## Usage Examples

### Example 1: Analyze a Namespace

```bash
curl -X POST http://localhost:8089/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "analyze",
      "arguments": {
        "namespace": "production",
        "explain": "true"
      }
    }
  }'
```

### Example 2: List Pods

```bash
curl -X POST http://localhost:8089/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "list-resources",
      "arguments": {
        "resourceType": "pods",
        "namespace": "default"
      }
    }
  }'
```

### Example 3: Get Pod Logs

```bash
curl -X POST http://localhost:8089/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "get-logs",
      "arguments": {
        "podName": "nginx-abc123",
        "namespace": "default",
        "tail": "100"
      }
    }
  }'
```

### Example 4: Access a Resource

```bash
curl -X POST http://localhost:8089/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "resources/read",
    "params": {
      "uri": "namespaces"
    }
  }'
```

### Example 5: Get a Troubleshooting Prompt

```bash
curl -X POST http://localhost:8089/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 5,
    "method": "prompts/get",
    "params": {
      "name": "troubleshoot-pod",
      "arguments": {
        "podName": "nginx-abc123",
        "namespace": "default"
      }
    }
  }'
```

## Integration with AI Assistants

### Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "k8sgpt": {
      "command": "k8sgpt",
      "args": ["serve", "--mcp"]
    }
  }
}
```

Restart Claude Desktop and you'll see k8sgpt tools available in the tool selector.

### Custom MCP Clients

Any MCP-compatible client can connect to the k8sgpt server. For HTTP-based clients:

1. Start the server: `k8sgpt serve --mcp --mcp-http --mcp-port 8089`
2. Connect to: `http://localhost:8089/mcp`
3. Use standard MCP protocol methods: `tools/list`, `tools/call`, `resources/read`, `prompts/get`

## HTTP API Reference

### Endpoint

```
POST http://localhost:8089/mcp
Content-Type: application/json
```

### Request Format

All requests follow the JSON-RPC 2.0 format:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "method_name",
  "params": {
    ...
  }
}
```

### Discovery Methods

**List Tools**
```json
{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}
```

**List Resources**
```json
{"jsonrpc": "2.0", "id": 2, "method": "resources/list"}
```

**List Prompts**
```json
{"jsonrpc": "2.0", "id": 3, "method": "prompts/list"}
```

### Tool Invocation

```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/call",
  "params": {
    "name": "tool_name",
    "arguments": {
      "arg1": "value1",
      "arg2": "value2"
    }
  }
}
```

### Resource Access

```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "resources/read",
  "params": {
    "uri": "resource_uri"
  }
}
```

### Prompt Access

```json
{
  "jsonrpc": "2.0",
  "id": 6,
  "method": "prompts/get",
  "params": {
    "name": "prompt_name",
    "arguments": {
      "arg1": "value1"
    }
  }
}
```

### Response Format

Successful responses:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    ...
  }
}
```

Error responses:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32600,
    "message": "Error description"
  }
}
```

## Advanced Configuration

### Custom Port

```bash
k8sgpt serve --mcp --mcp-http --mcp-port 9000
```

### With Specific Backend

```bash
k8sgpt serve --mcp --backend openai
```

### With Kubeconfig

```bash
k8sgpt serve --mcp --kubeconfig ~/.kube/config
```

## Troubleshooting

### Connection Issues

Verify the server is running:
```bash
curl http://localhost:8089/mcp
```

### Permission Issues

Ensure your kubeconfig has appropriate cluster access:
```bash
kubectl cluster-info
```

### Tool Errors

List available tools to verify names:
```bash
curl -X POST http://localhost:8089/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "id": 1, "method": "tools/list"}'
```

## Learn More

- [MCP Specification](https://modelcontextprotocol.io/)
- [K8sGPT Documentation](https://docs.k8sgpt.ai/)
- [MCP Go Library](https://github.com/mark3labs/mcp-go)
