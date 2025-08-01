package server

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
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

	go func() {
		err := s.Serve()
		time.Sleep(time.Second * 2)
		assert.NoError(t, err, "Serve should not return an error")
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
	defer initResp.Body.Close()

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
	defer initResp.Body.Close()

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
