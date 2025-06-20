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
	{
		Name:       "anthropic.claude-3-7-sonnet-20250219-v1:0",
		Completion: &bedrock_support.CohereCompletion{},
		Response:   &bedrock_support.CohereResponse{},
		Config: bedrock_support.BedrockModelConfig{
			MaxTokens:   100,
			Temperature: 0.5,
			TopP:        0.9,
			ModelName:   "anthropic.claude-3-7-sonnet-20250219-v1:0",
		},
	},
}

func TestBedrockModelConfig(t *testing.T) {
	client := &AmazonBedRockClient{models: testModels}

	// Should return error for ARN input (no exact match)
	_, err := client.getModelFromString("arn:aws:bedrock:us-east-1:*:inference-policy/anthropic.claude-3-5-sonnet-20240620-v1:0")
	assert.NotNil(t, err, "Should return error for ARN input")
}

func TestBedrockInvalidModel(t *testing.T) {
	client := &AmazonBedRockClient{models: testModels}

	// Should return error for invalid model name
	_, err := client.getModelFromString("arn:aws:s3:us-east-1:*:inference-policy/anthropic.claude-3-5-sonnet-20240620-v1:0")
	assert.NotNil(t, err, "Should return error for invalid model name")
}

func TestBedrockInferenceProfileARN(t *testing.T) {
	// Create a mock client with test models
	client := &AmazonBedRockClient{models: testModels}

	// Test with a valid inference profile ARN
	inferenceProfileARN := "arn:aws:bedrock:us-east-1:123456789012:inference-profile/my-profile"
	config := AIProvider{
		Model:          inferenceProfileARN,
		ProviderRegion: "us-east-1",
	}

	// This will fail in a real environment without mocks, but we're just testing the validation logic
	err := client.Configure(&config)
	// We expect an error since we can't actually call AWS in tests
	assert.NotNil(t, err, "Error should not be nil without AWS mocks")

	// Test with a valid application inference profile ARN
	appInferenceProfileARN := "arn:aws:bedrock:us-east-1:123456789012:application-inference-profile/my-profile"
	config = AIProvider{
		Model:          appInferenceProfileARN,
		ProviderRegion: "us-east-1",
	}

	// This will fail in a real environment without mocks, but we're just testing the validation logic
	err = client.Configure(&config)
	// We expect an error since we can't actually call AWS in tests
	assert.NotNil(t, err, "Error should not be nil without AWS mocks")

	// Test with an invalid inference profile ARN format
	invalidARN := "arn:aws:bedrock:us-east-1:123456789012:invalid-resource/my-profile"
	config = AIProvider{
		Model:          invalidARN,
		ProviderRegion: "us-east-1",
	}

	err = client.Configure(&config)
	assert.NotNil(t, err, "Error should not be nil for invalid inference profile ARN format")
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

func TestValidateInferenceProfileArn(t *testing.T) {
	tests := []struct {
		name  string
		arn   string
		valid bool
	}{
		{
			name:  "valid inference profile ARN",
			arn:   "arn:aws:bedrock:us-east-1:123456789012:inference-profile/my-profile",
			valid: true,
		},
		{
			name:  "valid application inference profile ARN",
			arn:   "arn:aws:bedrock:us-east-1:123456789012:application-inference-profile/my-profile",
			valid: true,
		},
		{
			name:  "invalid service in ARN",
			arn:   "arn:aws:s3:us-east-1:123456789012:inference-profile/my-profile",
			valid: false,
		},
		{
			name:  "invalid resource type in ARN",
			arn:   "arn:aws:bedrock:us-east-1:123456789012:model/my-profile",
			valid: false,
		},
		{
			name:  "malformed ARN",
			arn:   "arn:aws:bedrock:us-east-1:inference-profile/my-profile",
			valid: false,
		},
		{
			name:  "not an ARN",
			arn:   "not-an-arn",
			valid: false,
		},
		{
			name:  "empty string",
			arn:   "",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateInferenceProfileArn(tt.arn)
			assert.Equal(t, tt.valid, result, "validateInferenceProfileArn() result should match expected")
		})
	}
}
