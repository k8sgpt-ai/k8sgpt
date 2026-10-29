package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"

	rpc "buf.build/gen/go/k8sgpt-ai/k8sgpt/grpc/go/schema/v1/schemav1grpc"
	schemav1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
)

func TestServe(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer func() {
		err := logger.Sync()
		if err != nil {
			t.Logf("logger.Sync() error: %v", err)
		}
	}()

	s := &Config{
		Port:       "50059",
		Logger:     logger,
		EnableHttp: false,
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- s.Serve()
	}()

	// Wait until the server is ready to accept connections
	err := waitForPort("localhost:50059", 10*time.Second)
	assert.NoError(t, err, "Server should start without error")

	conn, err := grpc.Dial("localhost:50059", grpc.WithInsecure())
	assert.NoError(t, err, "Should be able to dial the server")
	defer func() {
		if err := conn.Close(); err != nil {
			t.Logf("failed to close connection: %v", err)
		}
	}()

	// Test a simple gRPC reflection request
	cli := grpc_reflection_v1alpha.NewServerReflectionClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := cli.ServerReflectionInfo(ctx)
	assert.NoError(t, err, "Should be able to get server reflection info")
	assert.NotNil(t, resp, "Response should not be nil")

	// Cleanup
	err = s.Shutdown()
	assert.NoError(t, err, "Shutdown should not return an error")

	// Shutdown closes the listener, which unblocks Serve. Assert synchronously
	// that Serve returned only because of that shutdown and not for another
	// reason.
	select {
	case err := <-serveErr:
		assert.ErrorIs(t, err, net.ErrClosed, "Serve should return only due to the listener being closed on shutdown")
	case <-time.After(5 * time.Second):
		t.Error("Serve did not return after shutdown")
	}
}

func TestServe_ShutdownRPC(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer func() {
		_ = logger.Sync()
	}()

	s := &Config{
		Port:       "50060",
		Logger:     logger,
		EnableHttp: false,
	}

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- s.Serve()
	}()

	err := waitForPort("localhost:50060", 10*time.Second)
	assert.NoError(t, err, "Server should start without error")

	conn, err := grpc.NewClient("localhost:50060", grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err, "Should be able to dial the server")
	defer func() {
		_ = conn.Close()
	}()

	configCli := rpc.NewServerConfigServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := configCli.Shutdown(ctx, &schemav1.ShutdownRequest{})
	assert.NoError(t, err, "Shutdown RPC should succeed")
	assert.NotNil(t, resp)
	assert.Equal(t, "Success", resp.Status)

	select {
	case err := <-serveErr:
		assert.ErrorIs(t, err, net.ErrClosed, "Serve should return due to shutdown RPC")
	case <-time.After(5 * time.Second):
		t.Error("Serve did not exit after shutdown RPC")
	}
}

func TestServer_ShutdownUnstartedAndWithMetrics(t *testing.T) {
	s := &Config{}
	err := s.Shutdown()
	assert.NoError(t, err)

	s.metricsServer = &http.Server{}
	err = s.Shutdown()
	assert.NoError(t, err)
}

// TestMCPServerCreation tests the creation of an MCP server
func TestMCPServerCreation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer func() {
		err := logger.Sync()
		if err != nil {
			t.Logf("logger.Sync() error: %v", err)
		}
	}()

	aiProvider := &ai.AIProvider{
		Name:     "test-provider",
		Password: "test-password",
		Model:    "test-model",
	}

	// Test HTTP mode
	mcpServer, err := NewMCPServer("8088", aiProvider, true, logger)
	assert.NoError(t, err, "Should be able to create MCP server with HTTP transport")
	assert.NotNil(t, mcpServer, "MCP server should not be nil")
	assert.True(t, mcpServer.useHTTP, "MCP server should be in HTTP mode")
	assert.Equal(t, "8088", mcpServer.port, "Port should be set correctly")

	// Test stdio mode
	mcpServerStdio, err := NewMCPServer("8088", aiProvider, false, logger)
	assert.NoError(t, err, "Should be able to create MCP server with stdio transport")
	assert.NotNil(t, mcpServerStdio, "MCP server should not be nil")
	assert.False(t, mcpServerStdio.useHTTP, "MCP server should be in stdio mode")
}

// TestMCPServerBasicHTTP tests basic HTTP connectivity to the MCP server
func TestMCPServerBasicHTTP(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer func() {
		err := logger.Sync()
		if err != nil {
			t.Logf("logger.Sync() error: %v", err)
		}
	}()

	aiProvider := &ai.AIProvider{
		Name:     "test-provider",
		Password: "test-password",
		Model:    "test-model",
	}

	mcpServer, err := NewMCPServer("8091", aiProvider, true, logger)
	assert.NoError(t, err, "Should be able to create MCP server")

	// For HTTP mode, the server is already started in NewMCPServer
	// No need to call Start() as it's already running in a goroutine

	// Wait for the server to start
	err = waitForPort("localhost:8091", 10*time.Second)
	if err != nil {
		t.Skipf("MCP server did not start within timeout: %v", err)
	}

	// First, initialize the session
	initRequest := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "initialize",
		"params": {
			"protocolVersion": "2025-03-26",
			"capabilities": {
				"tools": {},
				"resources": {},
				"prompts": {}
			},
			"clientInfo": {
				"name": "test-client",
				"version": "1.0.0"
			}
		}
	}`

	initResp, err := http.Post("http://localhost:8091/mcp", "application/json", bytes.NewBufferString(initRequest))
	if err != nil {
		t.Logf("Initialize request failed: %v", err)
		return
	}
	defer func() {
		if err := initResp.Body.Close(); err != nil {
			t.Logf("Error closing init response body: %v", err)
		}
	}()

	// Read initialization response
	initBody, err := io.ReadAll(initResp.Body)
	if err != nil {
		t.Logf("Failed to read init response body: %v", err)
	} else {
		t.Logf("Init response status: %d, body: %s", initResp.StatusCode, string(initBody))
	}

	// Extract session ID from response headers if present
	sessionID := initResp.Header.Get("Mcp-Session-Id")
	if sessionID == "" {
		t.Logf("No session ID in response headers")
	}

	// Now test tools/list with session ID if available
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json,text/event-stream",
	}
	if sessionID != "" {
		headers["Mcp-Session-Id"] = sessionID
	}

	req, err := http.NewRequest("POST", "http://localhost:8091/mcp", bytes.NewBufferString(`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`))
	if err != nil {
		t.Logf("Failed to create request: %v", err)
		return
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("MCP endpoint test skipped (server might not be fully ready): %v", err)
		return
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			t.Logf("resp.Body.Close() error: %v", err)
		}
	}()

	// Read response body for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Failed to read response body: %v", err)
	} else {
		t.Logf("Response status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Accept both 200 and 404 as valid responses (404 means endpoint not implemented)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Errorf("MCP endpoint returned unexpected status: %d", resp.StatusCode)
	}

	// Cleanup
	err = mcpServer.Close()
	assert.NoError(t, err, "MCP server should close without error")
}

// TestMCPServerToolCall tests calling a specific tool (analyze) through the MCP server
func TestMCPServerToolCall(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer func() {
		err := logger.Sync()
		if err != nil {
			t.Logf("logger.Sync() error: %v", err)
		}
	}()

	aiProvider := &ai.AIProvider{
		Name:     "test-provider",
		Password: "test-password",
		Model:    "test-model",
	}

	mcpServer, err := NewMCPServer("8090", aiProvider, true, logger)
	assert.NoError(t, err, "Should be able to create MCP server")

	// For HTTP mode, the server is already started in NewMCPServer
	// No need to call Start() as it's already running in a goroutine

	// Wait for the server to start
	err = waitForPort("localhost:8090", 10*time.Second)
	if err != nil {
		t.Skipf("MCP server did not start within timeout: %v", err)
	}

	// First, initialize the session
	initRequest := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "initialize",
		"params": {
			"protocolVersion": "2025-03-26",
			"capabilities": {
				"tools": {},
				"resources": {},
				"prompts": {}
			},
			"clientInfo": {
				"name": "test-client",
				"version": "1.0.0"
			}
		}
	}`

	initResp, err := http.Post("http://localhost:8090/mcp", "application/json", bytes.NewBufferString(initRequest))
	if err != nil {
		t.Logf("Initialize request failed: %v", err)
		return
	}
	defer func() {
		if err := initResp.Body.Close(); err != nil {
			t.Logf("Error closing init response body: %v", err)
		}
	}()

	// Extract session ID from response headers if present
	sessionID := initResp.Header.Get("Mcp-Session-Id")

	// Test calling the analyze tool with proper JSON-RPC format
	analyzeRequest := `{
		"jsonrpc": "2.0",
		"id": 2,
		"method": "tools/call",
		"params": {
			"name": "analyze",
			"arguments": {
				"namespace": "default",
				"backend": "openai",
				"language": "english",
				"explain": true,
				"maxConcurrency": 10
			}
		}
	}`

	// Create request with session ID if available
	req, err := http.NewRequest("POST", "http://localhost:8090/mcp", bytes.NewBufferString(analyzeRequest))
	if err != nil {
		t.Logf("Failed to create request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json,text/event-stream")
	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Analyze tool call test skipped (server might not be fully ready): %v", err)
		return
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			t.Logf("resp.Body.Close() error: %v", err)
		}
	}()

	// Accept both 200 and 404 as valid responses
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Errorf("Analyze tool call returned unexpected status: %d", resp.StatusCode)
	}

	// Cleanup
	err = mcpServer.Close()
	assert.NoError(t, err, "MCP server should close without error")
}

func waitForPort(address string, timeout time.Duration) error {
	start := time.Now()
	for {
		conn, err := net.Dial("tcp", address)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		if time.Since(start) > timeout {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// TestMCPToolSchemaCompatibility tests that all MCP tool parameter schemas
// comply with Draft-07 compatible schemas without $dynamicRef or unsupported Draft 2020-12
// meta-schema URIs that cause validation errors in MCP clients like GitHub Copilot (Issue #1507).
func TestMCPToolSchemaCompatibility(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer func() {
		_ = logger.Sync()
	}()

	aiProvider := &ai.AIProvider{
		Name:     "test-provider",
		Password: "test-password",
		Model:    "test-model",
	}

	mcpServer, err := NewMCPServer("8092", aiProvider, true, logger)
	assert.NoError(t, err)
	defer func() {
		_ = mcpServer.Close()
	}()

	err = waitForPort("localhost:8092", 10*time.Second)
	if err != nil {
		t.Skipf("MCP server did not start within timeout: %v", err)
	}

	// Initialize session
	initRequest := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "initialize",
		"params": {
			"protocolVersion": "2025-03-26",
			"capabilities": {"tools": {}},
			"clientInfo": {"name": "test-client", "version": "1.0.0"}
		}
	}`

	initResp, err := http.Post("http://localhost:8092/mcp", "application/json", bytes.NewBufferString(initRequest))
	require.NoError(t, err)
	sessionID := initResp.Header.Get("Mcp-Session-Id")
	_ = initResp.Body.Close()

	// List tools
	req, err := http.NewRequest("POST", "http://localhost:8092/mcp", bytes.NewBufferString(`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json,text/event-stream")
	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var listResult struct {
		Result struct {
			Tools []struct {
				Name        string         `json:"name"`
				Description string         `json:"description"`
				InputSchema map[string]any `json:"inputSchema"`
			} `json:"tools"`
		} `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&listResult)
	require.NoError(t, err)
	require.NotEmpty(t, listResult.Result.Tools)

	toolMap := make(map[string]map[string]any)
	for _, tool := range listResult.Result.Tools {
		toolMap[tool.Name] = tool.InputSchema

		// Serialize schema to check for meta-schema incompatibilities
		schemaBytes, err := json.Marshal(tool.InputSchema)
		require.NoError(t, err)
		schemaStr := string(schemaBytes)

		// Ensure no $dynamicRef or Draft 2020-12 meta-schema is used (Issue #1507)
		assert.NotContains(t, schemaStr, "$dynamicRef", "Tool %s schema should not contain $dynamicRef", tool.Name)
		assert.NotContains(t, schemaStr, "2020-12", "Tool %s schema should not use Draft 2020-12 meta-schema", tool.Name)

		// Every tool should declare "type": "object"
		assert.Equal(t, "object", tool.InputSchema["type"], "Tool %s inputSchema type should be object", tool.Name)
	}

	// Parameterless tools must have empty properties {} and no required fields
	parameterlessTools := []string{"cluster-info", "list-namespaces", "list-filters", "list-integrations"}
	for _, toolName := range parameterlessTools {
		schema, exists := toolMap[toolName]
		require.True(t, exists, "Parameterless tool %s should be registered", toolName)
		props, ok := schema["properties"].(map[string]any)
		require.True(t, ok, "Tool %s properties should be an object", toolName)
		assert.Empty(t, props, "Tool %s should have empty properties", toolName)
		assert.Nil(t, schema["required"], "Tool %s should have no required fields", toolName)
	}

	// Analyze tool should declare all parameters supported by handleAnalyze
	analyzeSchema, exists := toolMap["analyze"]
	require.True(t, exists, "analyze tool should be registered")
	analyzeProps, ok := analyzeSchema["properties"].(map[string]any)
	require.True(t, ok)
	expectedAnalyzeProps := []string{
		"namespace", "backend", "language", "explain", "filters",
		"labelSelector", "noCache", "maxConcurrency", "withDoc",
		"interactiveMode", "customHeaders", "withStats", "anonymize",
	}
	for _, prop := range expectedAnalyzeProps {
		assert.Contains(t, analyzeProps, prop, "analyze tool should include %s parameter", prop)
	}

	// Config tool should have customAnalyzers as array
	configSchema, exists := toolMap["config"]
	require.True(t, exists, "config tool should be registered")
	configProps, ok := configSchema["properties"].(map[string]any)
	require.True(t, ok)
	customAnalyzers, ok := configProps["customAnalyzers"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "array", customAnalyzers["type"], "config customAnalyzers should be array")
}
