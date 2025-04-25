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

// AnalyzeResponse represents the output of the analyze tool
type AnalyzeResponse struct {
	Content []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"content"`
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

	// Convert request to JSON
	reqJSON, err := json.Marshal(req)
	if err != nil {
		log.Fatalf("Failed to marshal request: %v", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	// Send request to MCP server
	resp, err := client.Post(
		fmt.Sprintf("http://localhost:%s/mcp/analyze", *serverPort),
		"application/json",
		bytes.NewBuffer(reqJSON),
	)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read and print raw response for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}
	fmt.Printf("Raw response: %s\n", string(body))

	// Parse response
	var analyzeResp AnalyzeResponse
	if err := json.Unmarshal(body, &analyzeResp); err != nil {
		log.Fatalf("Failed to decode response: %v", err)
	}

	// Print results
	fmt.Println("Analysis Results:")
	if len(analyzeResp.Content) > 0 {
		fmt.Println(analyzeResp.Content[0].Text)
	} else {
		fmt.Println("No results returned")
	}
}
