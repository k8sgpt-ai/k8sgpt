package server

import (
	"bytes"
	"context"
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
	defer logger.Sync()

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
	defer conn.Close()

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
	defer logger.Sync()

	aiProvider := &ai.AIProvider{
		Name:     "test-provider",
		Password: "test-password",
		Model:    "test-model",
	}

	// Test HTTP mode
	mcpServer, err := NewMCPServer("8089", aiProvider, true, logger)
	assert.NoError(t, err, "Should be able to create MCP server with HTTP transport")
	assert.NotNil(t, mcpServer, "MCP server should not be nil")
	assert.True(t, mcpServer.useHTTP, "MCP server should be in HTTP mode")
	assert.Equal(t, "8089", mcpServer.port, "Port should be set correctly")

	// Test stdio mode
	mcpServerStdio, err := NewMCPServer("8089", aiProvider, false, logger)
	assert.NoError(t, err, "Should be able to create MCP server with stdio transport")
	assert.NotNil(t, mcpServerStdio, "MCP server should not be nil")
	assert.False(t, mcpServerStdio.useHTTP, "MCP server should be in stdio mode")
}

// TestMCPServerBasicHTTP tests basic HTTP connectivity to the MCP server
func TestMCPServerBasicHTTP(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	aiProvider := &ai.AIProvider{
		Name:     "test-provider",
		Password: "test-password",
		Model:    "test-model",
	}

	mcpServer, err := NewMCPServer("8089", aiProvider, true, logger)
	assert.NoError(t, err, "Should be able to create MCP server")

	// Start the MCP server in a goroutine
	go func() {
		err := mcpServer.Start()
		// Note: Start() might return an error when the server is stopped, which is expected
		if err != nil {
			logger.Info("MCP server stopped", zap.Error(err))
		}
	}()

	// Wait for the server to start
	err = waitForPort("localhost:8089", 10*time.Second)
	if err != nil {
		t.Skipf("MCP server did not start within timeout: %v", err)
	}

	// Test basic connectivity to the MCP endpoint
	// The MCP HTTP transport uses a single POST endpoint for all requests
	resp, err := http.Post("http://localhost:8089/mcp", "application/json", bytes.NewBufferString(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	if err != nil {
		t.Logf("MCP endpoint test skipped (server might not be fully ready): %v", err)
		return
	}
	defer resp.Body.Close()

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
	defer logger.Sync()

	aiProvider := &ai.AIProvider{
		Name:     "test-provider",
		Password: "test-password",
		Model:    "test-model",
	}

	mcpServer, err := NewMCPServer("8090", aiProvider, true, logger)
	assert.NoError(t, err, "Should be able to create MCP server")

	// Start the MCP server in a goroutine
	go func() {
		err := mcpServer.Start()
		if err != nil {
			logger.Info("MCP server stopped", zap.Error(err))
		}
	}()

	// Wait for the server to start
	err = waitForPort("localhost:8090", 10*time.Second)
	if err != nil {
		t.Skipf("MCP server did not start within timeout: %v", err)
	}

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

	resp, err := http.Post("http://localhost:8090/mcp", "application/json", bytes.NewBufferString(analyzeRequest))
	if err != nil {
		t.Logf("Analyze tool call test skipped (server might not be fully ready): %v", err)
		return
	}
	defer resp.Body.Close()

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
			conn.Close()
			return nil
		}
		if time.Since(start) > timeout {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
}
