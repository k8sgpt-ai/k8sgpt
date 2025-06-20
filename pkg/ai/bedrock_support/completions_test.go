package bedrock_support

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCohereCompletion_GetCompletion(t *testing.T) {
	completion := &CohereCompletion{}
	modelConfig := BedrockModelConfig{
		MaxTokens:   100,
		Temperature: 0.7,
		TopP:        0.9,
	}
	prompt := "Test prompt"

	body, err := completion.GetCompletion(context.Background(), prompt, modelConfig)
	assert.NoError(t, err)

	var request map[string]interface{}
	err = json.Unmarshal(body, &request)
	assert.NoError(t, err)

	assert.Equal(t, "\n\nHuman: Test prompt  \n\nAssistant:", request["prompt"])
	assert.Equal(t, 100, int(request["max_tokens_to_sample"].(float64)))
	assert.Equal(t, 0.7, request["temperature"])
	assert.Equal(t, 0.9, request["top_p"])
}

func TestAI21_GetCompletion(t *testing.T) {
	completion := &AI21{}
	modelConfig := BedrockModelConfig{
		MaxTokens:   150,
		Temperature: 0.6,
		TopP:        0.8,
	}
	prompt := "Another test prompt"

	body, err := completion.GetCompletion(context.Background(), prompt, modelConfig)
	assert.NoError(t, err)

	var request map[string]interface{}
	err = json.Unmarshal(body, &request)
	assert.NoError(t, err)

	assert.Equal(t, "Another test prompt", request["prompt"])
	assert.Equal(t, 150, int(request["maxTokens"].(float64)))
	assert.Equal(t, 0.6, request["temperature"])
	assert.Equal(t, 0.8, request["topP"])
}

func TestAmazonCompletion_GetDefaultCompletion(t *testing.T) {
	completion := &AmazonCompletion{}
	modelConfig := BedrockModelConfig{
		MaxTokens:   200,
		Temperature: 0.5,
		TopP:        0.7,
		ModelName:   "amazon.titan-text-express-v1",
	}
	prompt := "Default test prompt"

	body, err := completion.GetDefaultCompletion(context.Background(), prompt, modelConfig)
	assert.NoError(t, err)

	var request map[string]interface{}
	err = json.Unmarshal(body, &request)
	assert.NoError(t, err)

	assert.Equal(t, "\n\nUser: Default test prompt", request["inputText"])
	textConfig := request["textGenerationConfig"].(map[string]interface{})
	assert.Equal(t, 200, int(textConfig["maxTokenCount"].(float64)))
	assert.Equal(t, 0.5, textConfig["temperature"])
	assert.Equal(t, 0.7, textConfig["topP"])
}

func TestAmazonCompletion_GetNovaCompletion(t *testing.T) {
	completion := &AmazonCompletion{}
	modelConfig := BedrockModelConfig{
		MaxTokens:   250,
		Temperature: 0.4,
		TopP:        0.6,
		ModelName:   "amazon.nova-pro-v1:0",
	}
	prompt := "Nova test prompt"

	body, err := completion.GetNovaCompletion(context.Background(), prompt, modelConfig)
	assert.NoError(t, err)

	var request map[string]interface{}
	err = json.Unmarshal(body, &request)
	assert.NoError(t, err)

	inferenceConfig := request["inferenceConfig"].(map[string]interface{})
	assert.Equal(t, 250, int(inferenceConfig["max_new_tokens"].(float64)))
	assert.Equal(t, 0.4, inferenceConfig["temperature"])
	assert.Equal(t, 0.6, inferenceConfig["topP"])

	messages := request["messages"].([]interface{})
	message := messages[0].(map[string]interface{})
	content := message["content"].([]interface{})
	contentMap := content[0].(map[string]interface{})
	assert.Equal(t, "Nova test prompt", contentMap["text"])
}

func TestAmazonCompletion_GetCompletion_Nova(t *testing.T) {
	completion := &AmazonCompletion{}
	modelConfig := BedrockModelConfig{
		MaxTokens:   250,
		Temperature: 0.4,
		TopP:        0.6,
		ModelName:   "amazon.nova-pro-v1:0",
	}
	prompt := "Nova test prompt"

	body, err := completion.GetCompletion(context.Background(), prompt, modelConfig)
	assert.NoError(t, err)

	var request map[string]interface{}
	err = json.Unmarshal(body, &request)
	assert.NoError(t, err)

	inferenceConfig := request["inferenceConfig"].(map[string]interface{})
	assert.Equal(t, 250, int(inferenceConfig["max_new_tokens"].(float64)))
	assert.Equal(t, 0.4, inferenceConfig["temperature"])
	assert.Equal(t, 0.6, inferenceConfig["topP"])

	messages := request["messages"].([]interface{})
	message := messages[0].(map[string]interface{})
	content := message["content"].([]interface{})
	contentMap := content[0].(map[string]interface{})
	assert.Equal(t, "Nova test prompt", contentMap["text"])
}

func TestAmazonCompletion_GetCompletion_Default(t *testing.T) {
	completion := &AmazonCompletion{}
	modelConfig := BedrockModelConfig{
		MaxTokens:   200,
		Temperature: 0.5,
		TopP:        0.7,
		ModelName:   "amazon.titan-text-express-v1",
	}
	prompt := "Default test prompt"

	body, err := completion.GetCompletion(context.Background(), prompt, modelConfig)
	assert.NoError(t, err)

	var request map[string]interface{}
	err = json.Unmarshal(body, &request)
	assert.NoError(t, err)

	assert.Equal(t, "\n\nUser: Default test prompt", request["inputText"])
	textConfig := request["textGenerationConfig"].(map[string]interface{})
	assert.Equal(t, 200, int(textConfig["maxTokenCount"].(float64)))
	assert.Equal(t, 0.5, textConfig["temperature"])
	assert.Equal(t, 0.7, textConfig["topP"])
}

func TestAmazonCompletion_GetCompletion_UnsupportedModel(t *testing.T) {
	completion := &AmazonCompletion{}
	modelConfig := BedrockModelConfig{
		MaxTokens:   200,
		Temperature: 0.5,
		TopP:        0.7,
		ModelName:   "unsupported-model",
	}
	prompt := "Test prompt"

	_, err := completion.GetCompletion(context.Background(), prompt, modelConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model unsupported-model is not supported")
}

func TestAmazonCompletion_GetCompletion_Inference_Profile(t *testing.T) {
	completion := &AmazonCompletion{}
	modelConfig := BedrockModelConfig{
		MaxTokens:   200,
		Temperature: 0.5,
		TopP:        0.7,
		ModelName:   "arn:aws:bedrock:us-east-1:*:inference-policy/anthropic.claude-3-5-sonnet-20240620-v1:0",
	}
	prompt := "Test prompt"

	_, err := completion.GetCompletion(context.Background(), prompt, modelConfig)
	assert.NoError(t, err)
}

func TestIsModelSupported(t *testing.T) {
	supported := []string{
		"anthropic.claude-v2",
		"anthropic.claude-v1",
	}
	assert.True(t, IsModelSupported("anthropic.claude-v2", supported))
	assert.False(t, IsModelSupported("unsupported-model", supported))
}
