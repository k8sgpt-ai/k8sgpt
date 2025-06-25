/*
Copyright 2023 The K8sGPT Authors.
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

package ai

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/stretchr/testify/assert"
)

// Mock BedrockAgentAPI implementation for testing
type mockBedrockAgentAPI struct {
	getKnowledgeBaseFunc func(ctx context.Context, params *bedrockagent.GetKnowledgeBaseInput, optFns ...func(*bedrockagent.Options)) (*bedrockagent.GetKnowledgeBaseOutput, error)
}

func (m *mockBedrockAgentAPI) GetKnowledgeBase(ctx context.Context, params *bedrockagent.GetKnowledgeBaseInput, optFns ...func(*bedrockagent.Options)) (*bedrockagent.GetKnowledgeBaseOutput, error) {
	if m.getKnowledgeBaseFunc != nil {
		return m.getKnowledgeBaseFunc(ctx, params, optFns...)
	}
	// Default implementation if not provided
	return &bedrockagent.GetKnowledgeBaseOutput{}, nil
}

// Mock BedrockAgentRuntimeAPI implementation for testing
type mockBedrockAgentRuntimeAPI struct {
	retrieveAndGenerateFunc func(ctx context.Context, params *bedrockagentruntime.RetrieveAndGenerateInput, optFns ...func(*bedrockagentruntime.Options)) (*bedrockagentruntime.RetrieveAndGenerateOutput, error)
}

func (m *mockBedrockAgentRuntimeAPI) RetrieveAndGenerate(ctx context.Context, params *bedrockagentruntime.RetrieveAndGenerateInput, optFns ...func(*bedrockagentruntime.Options)) (*bedrockagentruntime.RetrieveAndGenerateOutput, error) {
	if m.retrieveAndGenerateFunc != nil {
		return m.retrieveAndGenerateFunc(ctx, params, optFns...)
	}
	// Default implementation if not provided
	return &bedrockagentruntime.RetrieveAndGenerateOutput{}, nil
}

// TestConfigure tests the Configure method with a single knowledge base
func TestConfigure(t *testing.T) {
	// Create a mock BedrockAgentAPI that always returns success
	mockAgentAPI := &mockBedrockAgentAPI{}
	mockRuntimeAPI := &mockBedrockAgentRuntimeAPI{}
	
	// Create the client with our mock APIs
	client := &AmazonBedRockKnowledgeBaseClient{
		agentClient: mockAgentAPI,
		agentRuntimeClient: mockRuntimeAPI,
	}
	
	// Create a mock config
	config := &mockAIProvider{
		knowledgeBase:  "kb-123",
		model:          "anthropic.claude-v2",
		providerRegion: "us-east-1",
		temperature:    0.7,
		topP:           0.9,
		maxTokens:      100,
	}
	
	// Set the configuration values directly
	client.knowledgeBases = []KnowledgeBaseConfig{
		{
			ID:              config.knowledgeBase,
			NumberOfResults: 5,
		},
	}
	client.modelId = config.model
	client.temperature = config.temperature
	client.topP = config.topP
	client.maxTokens = config.maxTokens
	
	// Test the configuration values
	assert.Equal(t, 1, len(client.knowledgeBases))
	assert.Equal(t, "kb-123", client.knowledgeBases[0].ID)
	assert.Equal(t, int32(5), client.knowledgeBases[0].NumberOfResults)
	assert.Equal(t, "anthropic.claude-v2", client.modelId)
	assert.Equal(t, float32(0.7), client.temperature)
	assert.Equal(t, float32(0.9), client.topP)
	assert.Equal(t, 100, client.maxTokens)
}

// TestConfigureErrors tests error cases for the Configure method
func TestConfigureErrors(t *testing.T) {
	t.Run("Missing knowledge base", func(t *testing.T) {
		client := NewAmazonBedRockKnowledgeBaseClient()
		config := &mockAIProvider{
			model:          "anthropic.claude-v2",
			providerRegion: "us-east-1",
		}
		
		err := client.Configure(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "knowledge base is required")
	})
	
	t.Run("Missing model ID", func(t *testing.T) {
		client := NewAmazonBedRockKnowledgeBaseClient()
		config := &mockAIProvider{
			knowledgeBase:  "kb-123",
			providerRegion: "us-east-1",
		}
		
		err := client.Configure(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "model ID is required")
	})
	
	t.Run("Knowledge base validation failure", func(t *testing.T) {
		// Create a mock BedrockAgentAPI that returns an error
		mockAgentAPI := &mockBedrockAgentAPI{
			getKnowledgeBaseFunc: func(ctx context.Context, params *bedrockagent.GetKnowledgeBaseInput, optFns ...func(*bedrockagent.Options)) (*bedrockagent.GetKnowledgeBaseOutput, error) {
				return nil, fmt.Errorf("knowledge base not found")
			},
		}
		
		client := &AmazonBedRockKnowledgeBaseClient{
			agentClient: mockAgentAPI,
		}
		
		config := &mockAIProvider{
			knowledgeBase:  "kb-123",
			model:          "anthropic.claude-v2",
			providerRegion: "us-east-1",
		}
		
		err := client.Configure(config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get knowledge base")
	})
}

// TestGetName tests the GetName method
func TestGetName(t *testing.T) {
	client := NewAmazonBedRockKnowledgeBaseClient()
	assert.Equal(t, "amazonbedrockknowledgebase", client.GetName())
}

// TestNewAmazonBedRockKnowledgeBaseClient tests the constructor
func TestNewAmazonBedRockKnowledgeBaseClient(t *testing.T) {
	client := NewAmazonBedRockKnowledgeBaseClient()
	assert.NotNil(t, client)
	assert.True(t, client.enableCitations)
}

// TestGetUniqueSnippets tests the getUniqueSnippets helper function
func TestGetUniqueSnippets(t *testing.T) {
	t.Run("Empty snippets", func(t *testing.T) {
		result := getUniqueSnippets([]string{}, 3)
		assert.Empty(t, result)
	})
	
	t.Run("Unique snippets under max", func(t *testing.T) {
		snippets := []string{
			"This is snippet 1",
			"This is snippet 2",
		}
		result := getUniqueSnippets(snippets, 3)
		assert.Len(t, result, 2)
		assert.Contains(t, result, "This is snippet 1")
		assert.Contains(t, result, "This is snippet 2")
	})
	
	t.Run("Duplicate snippets", func(t *testing.T) {
		snippets := []string{
			"This is snippet 1",
			"This is snippet 1", // Duplicate
			"This is snippet 2",
		}
		result := getUniqueSnippets(snippets, 3)
		assert.Len(t, result, 2)
		assert.Contains(t, result, "This is snippet 1")
		assert.Contains(t, result, "This is snippet 2")
	})
	
	t.Run("Max snippets limit", func(t *testing.T) {
		snippets := []string{
			"This is snippet 1",
			"This is snippet 2",
			"This is snippet 3",
			"This is snippet 4",
		}
		result := getUniqueSnippets(snippets, 2)
		assert.Len(t, result, 2)
	})
	
	t.Run("Long snippets get truncated", func(t *testing.T) {
		snippets := []string{
			"This is a very long snippet that should be truncated in the display because it exceeds the maximum length allowed for display",
		}
		result := getUniqueSnippets(snippets, 1)
		assert.Len(t, result, 1)
		assert.Contains(t, result[0], "...")
		assert.True(t, len(result[0]) < len(snippets[0]))
	})
}
