package ai

import (
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai/bedrock_support"
	"github.com/stretchr/testify/assert"
)

// Test models for unit testing
var testModels = []bedrock_support.BedrockModel{
	{
		Name:       "anthropic.claude-3-5-sonnet-20240620-v1:0",
		Completion: &bedrock_support.CohereMessagesCompletion{},
		Response:   &bedrock_support.CohereMessagesResponse{},
		Config: bedrock_support.BedrockModelConfig{
			MaxTokens:   100,
			Temperature: 0.5,
			TopP:        0.9,
			ModelName:   "anthropic.claude-3-5-sonnet-20240620-v1:0",
		},
	},
	{
		Name:       "anthropic.claude-3-5-sonnet-20241022-v2:0",
		Completion: &bedrock_support.CohereCompletion{},
		Response:   &bedrock_support.CohereResponse{},
		Config: bedrock_support.BedrockModelConfig{
			MaxTokens:   100,
			Temperature: 0.5,
			TopP:        0.9,
			ModelName:   "anthropic.claude-3-5-sonnet-20241022-v2:0",
		},
	},
}

func TestBedrockModelConfig(t *testing.T) {
	client := &AmazonBedRockClient{models: testModels}

	foundModel, err := client.getModelFromString("arn:aws:bedrock:us-east-1:*:inference-policy/anthropic.claude-3-5-sonnet-20240620-v1:0")
	assert.Nil(t, err, "Error should be nil")
	assert.Equal(t, foundModel.Config.MaxTokens, 100)
	assert.Equal(t, foundModel.Config.Temperature, float32(0.5))
	assert.Equal(t, foundModel.Config.TopP, float32(0.9))
	assert.Equal(t, foundModel.Config.ModelName, "arn:aws:bedrock:us-east-1:*:inference-policy/anthropic.claude-3-5-sonnet-20240620-v1:0")
}

func TestBedrockInvalidModel(t *testing.T) {
	client := &AmazonBedRockClient{models: testModels}

	foundModel, err := client.getModelFromString("arn:aws:s3:us-east-1:*:inference-policy/anthropic.claude-3-5-sonnet-20240620-v1:0")
	assert.Nil(t, err, "Error should be nil")
	assert.Equal(t, foundModel.Config.MaxTokens, 100)
}

func TestBedrockGetCompletionInferenceProfile(t *testing.T) {
	modelName := "arn:aws:bedrock:us-east-1:*:inference-policy/anthropic.claude-3-5-sonnet-20240620-v1:0"
	var inferenceModelModels = []bedrock_support.BedrockModel{
		{
			Name:       "anthropic.claude-3-5-sonnet-20240620-v1:0",
			Completion: &bedrock_support.CohereMessagesCompletion{},
			Response:   &bedrock_support.CohereMessagesResponse{},
			Config: bedrock_support.BedrockModelConfig{
				MaxTokens:   100,
				Temperature: 0.5,
				TopP:        0.9,
				ModelName:   modelName,
			},
		},
	}
	client := &AmazonBedRockClient{models: inferenceModelModels}

	config := AIProvider{
		Model: modelName,
	}
	err := client.Configure(&config)
	assert.Nil(t, err, "Error should be nil")
	assert.Equal(t, modelName, client.model.Config.ModelName, "Model name should match")
}

func TestGetModelFromString(t *testing.T) {
	client := &AmazonBedRockClient{models: testModels}

	tests := []struct {
		name      string
		model     string
		wantModel string
		wantErr   bool
	}{
		{
			name:      "exact model name match",
			model:     "anthropic.claude-3-5-sonnet-20240620-v1:0",
			wantModel: "anthropic.claude-3-5-sonnet-20240620-v1:0",
			wantErr:   false,
		},
		{
			name:      "partial model name match",
			model:     "claude-3-5-sonnet",
			wantModel: "anthropic.claude-3-5-sonnet-20240620-v1:0",
			wantErr:   false,
		},
		{
			name:      "model name with different version",
			model:     "anthropic.claude-3-5-sonnet-20241022-v2:0",
			wantModel: "anthropic.claude-3-5-sonnet-20241022-v2:0",
			wantErr:   false,
		},
		{
			name:      "non-existent model",
			model:     "non-existent-model",
			wantModel: "",
			wantErr:   true,
		},
		{
			name:      "empty model name",
			model:     "",
			wantModel: "",
			wantErr:   true,
		},
		{
			name:      "model name with extra spaces",
			model:     "  anthropic.claude-3-5-sonnet-20240620-v1:0  ",
			wantModel: "anthropic.claude-3-5-sonnet-20240620-v1:0",
			wantErr:   false,
		},
		{
			name:      "case insensitive match",
			model:     "ANTHROPIC.CLAUDE-3-5-SONNET-20240620-V1:0",
			wantModel: "anthropic.claude-3-5-sonnet-20240620-v1:0",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotModel, err := client.getModelFromString(tt.model)
			if (err != nil) != tt.wantErr {
				t.Errorf("getModelFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gotModel.Name != tt.wantModel {
				t.Errorf("getModelFromString() = %v, want %v", gotModel.Name, tt.wantModel)
			}
		})
	}
}

// TestDefaultModels tests that the client works with default models
func TestDefaultModels(t *testing.T) {
	client := &AmazonBedRockClient{}

	// Configure should initialize default models
	err := client.Configure(&AIProvider{
		Model: "anthropic.claude-v2",
	})

	assert.NoError(t, err, "Configure should not return an error")
	assert.NotNil(t, client.models, "Models should be initialized")
	assert.NotEmpty(t, client.models, "Models should not be empty")

	// Test finding a default model
	model, err := client.getModelFromString("anthropic.claude-v2")
	assert.NoError(t, err, "Should find the model")
	assert.Equal(t, "anthropic.claude-v2", model.Name, "Should find the correct model")
}
