package bedrock_support

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBedrockModelConfig(t *testing.T) {
	config := BedrockModelConfig{
		MaxTokens:   100,
		Temperature: 0.7,
		TopP:        0.9,
		ModelName:   "test-model",
	}

	assert.Equal(t, 100, config.MaxTokens)
	assert.Equal(t, float32(0.7), config.Temperature)
	assert.Equal(t, float32(0.9), config.TopP)
	assert.Equal(t, "test-model", config.ModelName)
}

func TestBedrockModel(t *testing.T) {
	completion := &MockCompletion{}
	response := &MockResponse{}
	config := BedrockModelConfig{
		MaxTokens:   100,
		Temperature: 0.7,
		TopP:        0.9,
		ModelName:   "test-model",
	}

	model := BedrockModel{
		Name:       "Test Model",
		Completion: completion,
		Response:   response,
		Config:     config,
	}

	assert.Equal(t, "Test Model", model.Name)
	assert.Equal(t, completion, model.Completion)
	assert.Equal(t, response, model.Response)
	assert.Equal(t, config, model.Config)
}

// MockCompletion is a mock implementation of the ICompletion interface
type MockCompletion struct{}

func (m *MockCompletion) GetCompletion(ctx context.Context, prompt string, config BedrockModelConfig) ([]byte, error) {
	return []byte(`{"prompt": "mock prompt"}`), nil
}

// MockResponse is a mock implementation of the IResponse interface
type MockResponse struct{}

func (m *MockResponse) ParseResponse(body []byte) (string, error) {
	return "mock response", nil
}
