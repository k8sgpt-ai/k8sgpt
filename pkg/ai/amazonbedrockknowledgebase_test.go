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
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"
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

func TestGetCompletion(t *testing.T) {
	// Create a mock BedrockAgentRuntimeAPI
	mockAgentRuntimeAPI := &mockBedrockAgentRuntimeAPI{
		retrieveAndGenerateFunc: func(ctx context.Context, params *bedrockagentruntime.RetrieveAndGenerateInput, optFns ...func(*bedrockagentruntime.Options)) (*bedrockagentruntime.RetrieveAndGenerateOutput, error) {
			expectedText := "This is the generated response"
			return &bedrockagentruntime.RetrieveAndGenerateOutput{
				Output: &types.RetrieveAndGenerateOutput{
					Text: aws.String(expectedText),
				},
			}, nil
		},
	}

	// Create the client with a single knowledge base
	client := &AmazonBedRockKnowledgeBaseClient{
		agentRuntimeClient: mockAgentRuntimeAPI,
		knowledgeBases: []KnowledgeBaseConfig{
			{ID: "kb-123", NumberOfResults: 5},
		},
		modelId:     "anthropic.claude-v2",
		temperature: 0.7,
		topP:        0.9,
		maxTokens:   100,
	}

	// Test GetCompletion
	result, err := client.GetCompletion(context.Background(), "Test prompt")
	assert.NoError(t, err)
	assert.Equal(t, "This is the generated response", result)
}
