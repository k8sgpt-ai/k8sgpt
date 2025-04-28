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
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/server"
	"go.uber.org/zap"
)

func main() {
	// Parse command line flags
	port := flag.String("port", "8089", "Port to run the MCP server on")
	useHTTP := flag.Bool("http", false, "Enable HTTP mode for MCP server")
	flag.Parse()

	// Initialize zap logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Error creating logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Error syncing logger: %v", err)
		}
	}()

	// Create AI provider
	aiProvider := &ai.AIProvider{
		Name:     "openai",
		Password: os.Getenv("OPENAI_API_KEY"),
		Model:    "gpt-3.5-turbo",
	}

	// Create and start MCP server
	mcpServer, err := server.NewMCPServer(*port, aiProvider, *useHTTP, logger)
	if err != nil {
		log.Fatalf("Error creating MCP server: %v", err)
	}

	// Start the server in a goroutine
	go func() {
		if err := mcpServer.Start(); err != nil {
			log.Fatalf("Error starting MCP server: %v", err)
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Cleanup
	if err := mcpServer.Close(); err != nil {
		log.Printf("Error closing MCP server: %v", err)
	}
}
