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
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

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

// JSONRPCResponse represents the JSON-RPC response format
type JSONRPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		Content []struct {
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"content"`
	} `json:"result,omitempty"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func main() {
	// Parse command line flags
	serverPort := flag.String("port", "8089", "Port of the MCP server")
	namespace := flag.String("namespace", "", "Kubernetes namespace to analyze")
	backend := flag.String("backend", "", "AI backend to use")
	language := flag.String("language", "english", "Language for analysis")
	flag.Parse()

	// Create analyze request
	req := AnalyzeRequest{
		Namespace:      *namespace,
		Backend:        *backend,
		Language:       *language,
		Explain:        true,
		MaxConcurrency: 10,
	}

	// Note: req is now used directly in the JSON-RPC request

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	// First, initialize the session
	initRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2025-03-26",
			"capabilities": map[string]interface{}{
				"tools":     map[string]interface{}{},
				"resources": map[string]interface{}{},
				"prompts":   map[string]interface{}{},
			},
			"clientInfo": map[string]interface{}{
				"name":    "k8sgpt-client",
				"version": "1.0.0",
			},
		},
	}

	initData, err := json.Marshal(initRequest)
	if err != nil {
		log.Fatalf("Failed to marshal init request: %v", err)
	}

	// Send initialization request
	initResp, err := client.Post(
		fmt.Sprintf("http://localhost:%s/mcp", *serverPort),
		"application/json",
		bytes.NewBuffer(initData),
	)
	if err != nil {
		log.Fatalf("Failed to send init request: %v", err)
	}
	defer func() {
		if err := initResp.Body.Close(); err != nil {
			log.Printf("Error closing init response body: %v", err)
		}
	}()

	// Extract session ID from response headers
	sessionID := initResp.Header.Get("Mcp-Session-Id")
	if sessionID == "" {
		log.Println("Warning: No session ID received from server")
	}

	// Create JSON-RPC request for analyze
	jsonRPCRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      "analyze",
			"arguments": req,
		},
	}

	// Convert to JSON
	jsonRPCData, err := json.Marshal(jsonRPCRequest)
	if err != nil {
		log.Fatalf("Failed to marshal JSON-RPC request: %v", err)
	}

	// Create request with session ID if available
	httpReq, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%s/mcp", *serverPort), bytes.NewBuffer(jsonRPCData))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json,text/event-stream")
	if sessionID != "" {
		httpReq.Header.Set("Mcp-Session-Id", sessionID)
	}

	// Send request to MCP server
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	// Read and print raw response for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}
	fmt.Printf("Raw response: %s\n", string(body))

	// Parse response
	var jsonRPCResp JSONRPCResponse
	if err := json.Unmarshal(body, &jsonRPCResp); err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}

	// Print results
	fmt.Println("Analysis Results:")
	if jsonRPCResp.Error != nil {
		fmt.Printf("Error: %s (code: %d)\n", jsonRPCResp.Error.Message, jsonRPCResp.Error.Code)
	} else if len(jsonRPCResp.Result.Content) > 0 {
		fmt.Println(jsonRPCResp.Result.Content[0].Text)
	} else {
		fmt.Println("No results returned")
	}
}
