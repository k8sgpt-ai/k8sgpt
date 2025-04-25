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

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analysis"
	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
)

// MCPServer represents an MCP server for k8sgpt
type MCPServer struct {
	server     *mcp_golang.Server
	port       string
	aiProvider *ai.AIProvider
	analysis   *analysis.Analysis
}

// NewMCPServer creates a new MCP server
func NewMCPServer(port string, aiProvider *ai.AIProvider) (*MCPServer, error) {
	// Create MCP server with stdio transport
	transport := stdio.NewStdioServerTransport()

	server := mcp_golang.NewServer(transport)

	// Initialize analysis configuration
	analysis, err := analysis.NewAnalysis(
		aiProvider.Name, // backend
		"english",       // language
		[]string{},      // filters
		"",              // namespace
		"",              // labelSelector
		false,           // nocache
		false,           // explain
		10,              // maxConcurrency
		false,           // withDoc
		false,           // interactiveMode
		[]string{},      // customHeaders
		false,           // withStats
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize analysis: %v", err)
	}

	return &MCPServer{
		server:     server,
		port:       port,
		aiProvider: aiProvider,
		analysis:   analysis,
	}, nil
}

// Start starts the MCP server
func (s *MCPServer) Start() error {
	if s.server == nil {
		return fmt.Errorf("server not initialized")
	}

	// Register analyze tool
	if err := s.server.RegisterTool("analyze", "Analyze Kubernetes resources", s.handleAnalyze); err != nil {
		return fmt.Errorf("failed to register analyze tool: %v", err)
	}

	// Register cluster info tool
	if err := s.server.RegisterTool("cluster-info", "Get Kubernetes cluster information", s.handleClusterInfo); err != nil {
		return fmt.Errorf("failed to register cluster-info tool: %v", err)
	}

	// Register resources
	if err := s.registerResources(); err != nil {
		return fmt.Errorf("failed to register resources: %v", err)
	}

	// Register prompts
	if err := s.registerPrompts(); err != nil {
		return fmt.Errorf("failed to register prompts: %v", err)
	}

	// Start the server
	return s.server.Serve()
}

// AnalyzeRequest represents the input parameters for the analyze tool
type AnalyzeRequest struct {
	Namespace       string   `json:"namespace,omitempty"`
	Backend         string   `json:"backend,omitempty"`
	Language        string   `json:"language,omitempty"`
	Filters         []string `json:"filters,omitempty"`
	LabelSelector   string   `json:"labelSelector,omitempty"`
	NoCache         bool     `json:"noCache,omitempty"`
	Explain         bool     `json:"explain,omitempty"`
	MaxConcurrency  int      `json:"maxConcurrency,omitempty"`
	WithDoc         bool     `json:"withDoc,omitempty"`
	InteractiveMode bool     `json:"interactiveMode,omitempty"`
	CustomHeaders   []string `json:"customHeaders,omitempty"`
	WithStats       bool     `json:"withStats,omitempty"`
}

// AnalyzeResponse represents the output of the analyze tool
type AnalyzeResponse struct {
	Results string `json:"results"`
}

// ClusterInfoRequest represents the input parameters for the cluster-info tool
type ClusterInfoRequest struct {
	// Empty struct as we don't need any input parameters
}

// ClusterInfoResponse represents the output of the cluster-info tool
type ClusterInfoResponse struct {
	Info string `json:"info"`
}

// handleAnalyze handles the analyze tool
func (s *MCPServer) handleAnalyze(ctx context.Context, request *AnalyzeRequest) (*mcp_golang.ToolResponse, error) {
	// Create a new analysis with the request parameters
	analysis, err := analysis.NewAnalysis(
		request.Backend,
		request.Language,
		request.Filters,
		request.Namespace,
		request.LabelSelector,
		request.NoCache,
		request.Explain,
		request.MaxConcurrency,
		request.WithDoc,
		request.InteractiveMode,
		request.CustomHeaders,
		request.WithStats,
	)
	if err != nil {
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Failed to create analysis: %v", err))), nil
	}
	defer analysis.Close()

	// Run the analysis
	analysis.RunAnalysis()

	// Get the output
	output, err := analysis.PrintOutput("json")
	if err != nil {
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Failed to print output: %v", err))), nil
	}

	return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(string(output))), nil
}

// handleClusterInfo handles the cluster-info tool
func (s *MCPServer) handleClusterInfo(ctx context.Context, request *ClusterInfoRequest) (*mcp_golang.ToolResponse, error) {
	// Get cluster info from the client
	version, err := s.analysis.Client.Client.Discovery().ServerVersion()
	if err != nil {
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("failed to get cluster version: %v", err))), nil
	}

	info := fmt.Sprintf("Kubernetes %s", version.GitVersion)
	return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(info)), nil
}

// registerPrompts registers the prompts for the MCP server
func (s *MCPServer) registerPrompts() error {
	// Register any prompts needed for the MCP server
	return nil
}

// registerResources registers the resources for the MCP server
func (s *MCPServer) registerResources() error {
	if err := s.server.RegisterResource("cluster-info", "Get cluster information", "Get information about the Kubernetes cluster", "text", s.getClusterInfo); err != nil {
		return fmt.Errorf("failed to register cluster-info resource: %v", err)
	}
	return nil
}

func (s *MCPServer) getClusterInfo(ctx context.Context) (interface{}, error) {
	// Get cluster info from the client
	version, err := s.analysis.Client.Client.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster version: %v", err)
	}

	return map[string]string{
		"version":    version.String(),
		"platform":   version.Platform,
		"gitVersion": version.GitVersion,
	}, nil
}

// Close closes the MCP server and releases resources
func (s *MCPServer) Close() error {
	if s.analysis != nil {
		s.analysis.Close()
	}
	return nil
}
