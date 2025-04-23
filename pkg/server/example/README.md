# K8sGPT MCP Server Example

This directory contains an example of how to use the K8sGPT MCP server in a real-world scenario.

## Prerequisites

- Go 1.16 or later
- Access to a Kubernetes cluster
- `kubectl` configured to access your cluster

## Building the Example

To build the example, run:

```bash
go build -o mcp-server-example
```

## Running the Example

To run the example, use the following command:

```bash
./mcp-server-example --kubeconfig=/path/to/kubeconfig --namespace=default
```

### Command-line Flags

- `--kubeconfig`: Path to the kubeconfig file (optional, defaults to the standard location)
- `--namespace`: Kubernetes namespace to analyze (optional)

## Example Output

When you run the example, you should see output similar to the following:

```
Starting MCP server...
```

The server will continue running until you press Ctrl+C to stop it.

## Integration with Mission Control

To integrate this example with Mission Control, you need to:

1. Start the MCP server using the example
2. Configure Mission Control to connect to the MCP server
3. Use Mission Control to analyze your Kubernetes cluster

## Troubleshooting

If you encounter any issues, check the following:

1. Ensure that your Kubernetes cluster is accessible
2. Verify that your kubeconfig file is valid
3. Check that the namespace you specified exists

## License

This code is licensed under the Apache License 2.0. 