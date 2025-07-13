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

	schemav1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analysis"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/k8sgpt-ai/k8sgpt/pkg/server/config"
	mcp_golang "github.com/metoro-io/mcp-golang"
	mcp_http "github.com/metoro-io/mcp-golang/transport/http"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// MCPServer represents an MCP server for k8sgpt
type MCPServer struct {
	server     *mcp_golang.Server
	port       string
	aiProvider *ai.AIProvider
	useHTTP    bool
	logger     *zap.Logger
}

// NewMCPServer creates a new MCP server
func NewMCPServer(port string, aiProvider *ai.AIProvider, useHTTP bool, logger *zap.Logger) (*MCPServer, error) {
	opts := []mcp_golang.ServerOptions{
		mcp_golang.WithName("k8sgpt"),
		mcp_golang.WithVersion("0.0.1"),
	}

	var server *mcp_golang.Server
	if useHTTP {
		logger.Info("starting MCP server with http transport on port", zap.String("port", port))
		httpTransport := mcp_http.NewHTTPTransport("/mcp").WithAddr(":" + port)
		server = mcp_golang.NewServer(httpTransport, opts...)
	} else {
		server = mcp_golang.NewServer(stdio.NewStdioServerTransport(), opts...)
	}

	return &MCPServer{
		server:     server,
		port:       port,
		aiProvider: aiProvider,
		useHTTP:    useHTTP,
		logger:     logger,
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

	// Register config tool
	if err := s.server.RegisterTool("config", "Configure K8sGPT settings", s.handleConfig); err != nil {
		return fmt.Errorf("failed to register config tool: %v", err)
	}

	// Register resources
	if err := s.registerResources(); err != nil {
		return fmt.Errorf("failed to register resources: %v", err)
	}

	// Register prompts
	if err := s.registerPrompts(); err != nil {
		return fmt.Errorf("failed to register prompts: %v", err)
	}

	// Start the server (this will block)
	if err := s.server.Serve(); err != nil {
		s.logger.Error("Error starting MCP server", zap.Error(err))
	}
	return nil
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

// ConfigRequest represents the input parameters for the config tool
type ConfigRequest struct {
	CustomAnalyzers []struct {
		Name       string `json:"name"`
		Connection struct {
			Url  string `json:"url"`
			Port int    `json:"port"`
		} `json:"connection"`
	} `json:"customAnalyzers,omitempty"`
	Cache struct {
		Type string `json:"type"`
		// S3 specific fields
		BucketName string `json:"bucketName,omitempty"`
		Region     string `json:"region,omitempty"`
		Endpoint   string `json:"endpoint,omitempty"`
		Insecure   bool   `json:"insecure,omitempty"`
		// Azure specific fields
		StorageAccount string `json:"storageAccount,omitempty"`
		ContainerName  string `json:"containerName,omitempty"`
		// GCS specific fields
		ProjectId string `json:"projectId,omitempty"`
	} `json:"cache,omitempty"`
}

// ConfigResponse represents the output of the config tool
type ConfigResponse struct {
	Status string `json:"status"`
}

// handleAnalyze handles the analyze tool
func (s *MCPServer) handleAnalyze(ctx context.Context, request *AnalyzeRequest) (*mcp_golang.ToolResponse, error) {
	// Get stored configuration
	var configAI ai.AIConfiguration
	if err := viper.UnmarshalKey("ai", &configAI); err != nil {
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Failed to load AI configuration: %v", err))), nil
	}
	// Use stored configuration if not specified in request
	if request.Backend == "" {
		if configAI.DefaultProvider != "" {
			request.Backend = configAI.DefaultProvider
		} else if len(configAI.Providers) > 0 {
			request.Backend = configAI.Providers[0].Name
		} else {
			request.Backend = "openai" // fallback default
		}
	}

	request.Explain = true
	// Get stored filters if not specified
	if len(request.Filters) == 0 {
		request.Filters = viper.GetStringSlice("active_filters")
	}

	// Validate MaxConcurrency to prevent excessive memory allocation
	request.MaxConcurrency = validateMaxConcurrency(request.MaxConcurrency)

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

// validateMaxConcurrency validates and bounds the MaxConcurrency parameter
func validateMaxConcurrency(maxConcurrency int) int {
	const maxAllowedConcurrency = 100
	if maxConcurrency <= 0 {
		return 10 // Default value if not set
	} else if maxConcurrency > maxAllowedConcurrency {
		return maxAllowedConcurrency // Cap at a reasonable maximum
	}
	return maxConcurrency
}

// handleClusterInfo handles the cluster-info tool
func (s *MCPServer) handleClusterInfo(ctx context.Context, request *ClusterInfoRequest) (*mcp_golang.ToolResponse, error) {
	// Create a new Kubernetes client
	client, err := kubernetes.NewClient("", "")
	if err != nil {
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("failed to create Kubernetes client: %v", err))), nil
	}

	// Get cluster info from the client
	version, err := client.Client.Discovery().ServerVersion()
	if err != nil {
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("failed to get cluster version: %v", err))), nil
	}

	info := fmt.Sprintf("Kubernetes %s", version.GitVersion)
	return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(info)), nil
}

// handleConfig handles the config tool
func (s *MCPServer) handleConfig(ctx context.Context, request *ConfigRequest) (*mcp_golang.ToolResponse, error) {
	// Create a new config handler
	handler := &config.Handler{}

	// Convert request to AddConfigRequest
	addConfigReq := &schemav1.AddConfigRequest{
		CustomAnalyzers: make([]*schemav1.CustomAnalyzer, 0),
	}

	// Add custom analyzers if present
	if len(request.CustomAnalyzers) > 0 {
		for _, ca := range request.CustomAnalyzers {
			addConfigReq.CustomAnalyzers = append(addConfigReq.CustomAnalyzers, &schemav1.CustomAnalyzer{
				Name: ca.Name,
				Connection: &schemav1.Connection{
					Url:  ca.Connection.Url,
					Port: fmt.Sprintf("%d", ca.Connection.Port),
				},
			})
		}
	}

	// Add cache configuration if present
	if request.Cache.Type != "" {
		cacheConfig := &schemav1.Cache{}
		switch request.Cache.Type {
		case "s3":
			cacheConfig.CacheType = &schemav1.Cache_S3Cache{
				S3Cache: &schemav1.S3Cache{
					BucketName: request.Cache.BucketName,
					Region:     request.Cache.Region,
					Endpoint:   request.Cache.Endpoint,
					Insecure:   request.Cache.Insecure,
				},
			}
		case "azure":
			cacheConfig.CacheType = &schemav1.Cache_AzureCache{
				AzureCache: &schemav1.AzureCache{
					StorageAccount: request.Cache.StorageAccount,
					ContainerName:  request.Cache.ContainerName,
				},
			}
		case "gcs":
			cacheConfig.CacheType = &schemav1.Cache_GcsCache{
				GcsCache: &schemav1.GCSCache{
					BucketName: request.Cache.BucketName,
					Region:     request.Cache.Region,
					ProjectId:  request.Cache.ProjectId,
				},
			}
		}
		addConfigReq.Cache = cacheConfig
	}

	// Apply the configuration using the shared function
	if err := handler.ApplyConfig(ctx, addConfigReq); err != nil {
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(fmt.Sprintf("Failed to add config: %v", err))), nil
	}

	return mcp_golang.NewToolResponse(mcp_golang.NewTextContent("Successfully added configuration")), nil
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

func (s *MCPServer) getClusterInfo(ctx context.Context) (*mcp_golang.ResourceResponse, error) {
	// Create a new Kubernetes client
	client, err := kubernetes.NewClient("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// Get cluster info from the client
	version, err := client.Client.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster version: %v", err)
	}

	data, err := json.Marshal(map[string]string{
		"version":    version.String(),
		"platform":   version.Platform,
		"gitVersion": version.GitVersion,
	})
	if err != nil {
		return mcp_golang.NewResourceResponse(
			mcp_golang.NewTextEmbeddedResource(
				"cluster-info",
				"Failed to marshal cluster info",
				"text/plain",
			),
		), nil
	}
	return mcp_golang.NewResourceResponse(
		mcp_golang.NewTextEmbeddedResource(
			"cluster-info",
			string(data),
			"application/json",
		),
	), nil
}

// Close closes the MCP server and releases resources
func (s *MCPServer) Close() error {
	return nil
}
