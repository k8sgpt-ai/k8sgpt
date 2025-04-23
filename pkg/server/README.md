# K8sGPT MCP Server

This directory contains the implementation of the Mission Control Protocol (MCP) server for K8sGPT. The MCP server allows K8sGPT to be integrated with other tools that support the MCP protocol.

## Components

- `mcp.go`: The main MCP server implementation
- `server.go`: The HTTP server implementation
- `tools.go`: Tool definitions for the MCP server

## Features

The MCP server provides the following features:

1. **Analyze Kubernetes Resources**: Analyze Kubernetes resources in a cluster
2. **Get Cluster Information**: Retrieve information about the Kubernetes cluster

## Usage

To use the MCP server, you need to:

1. Initialize the MCP server with a Kubernetes client
2. Start the server
3. Connect to the server using an MCP client

Example:

```go
client, err := kubernetes.NewForConfig(config)
if err != nil {
    log.Fatalf("Failed to create Kubernetes client: %v", err)
}

mcpServer := server.NewMCPServer(client)
if err := mcpServer.Start(); err != nil {
    log.Fatalf("Failed to start MCP server: %v", err)
}
```

## Integration

The MCP server can be integrated with other tools that support the MCP protocol, such as:

- Mission Control
- Other MCP-compatible tools

## License

This code is licensed under the Apache License 2.0.
