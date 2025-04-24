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

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
)

func main() {
	// Create transport and client
	transport := stdio.NewStdioServerTransport()

	// Create a new client with the transport
	client := mcp_golang.NewClient(transport)

	// Initialize the client
	if _, err := client.Initialize(context.Background()); err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	// Call analyze tool
	response, err := client.CallTool(context.Background(), "analyze", map[string]interface{}{
		"namespace": "default",
	})
	if err != nil {
		log.Fatalf("Failed to call analyze tool: %v", err)
	}

	// Print response
	if response != nil && len(response.Content) > 0 {
		fmt.Printf("Analysis results: %s\n", response.Content[0].TextContent.Text)
	}

	// Get cluster info
	clusterInfo, err := client.CallTool(context.Background(), "cluster-info", nil)
	if err != nil {
		log.Fatalf("Failed to get cluster info: %v", err)
	}

	// Print cluster info
	if clusterInfo != nil && len(clusterInfo.Content) > 0 {
		fmt.Printf("Cluster info: %s\n", clusterInfo.Content[0].TextContent.Text)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
