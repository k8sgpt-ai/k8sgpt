package bedrock_support

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type ICompletion interface {
	GetCompletion(ctx context.Context, prompt string, modelConfig BedrockModelConfig) ([]byte, error)
}

type CohereCompletion struct {
	completion ICompletion
}

func (a *CohereCompletion) GetCompletion(ctx context.Context, prompt string, modelConfig BedrockModelConfig) ([]byte, error) {
	request := map[string]interface{}{
		"prompt":               fmt.Sprintf("\n\nHuman: %s  \n\nAssistant:", prompt),
		"max_tokens_to_sample": modelConfig.MaxTokens,
		"temperature":          modelConfig.Temperature,
		"top_p":                modelConfig.TopP,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}

type CohereMessagesCompletion struct {
	completion ICompletion
}

func (a *CohereMessagesCompletion) GetCompletion(ctx context.Context, prompt string, modelConfig BedrockModelConfig) ([]byte, error) {
	request := map[string]interface{}{
		"max_tokens":        modelConfig.MaxTokens,
		"temperature":       modelConfig.Temperature,
		"top_p":             modelConfig.TopP,
		"anthropic_version": "bedrock-2023-05-31", // Or another valid version
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}

type AI21 struct {
	completion ICompletion
}

func (a *AI21) GetCompletion(ctx context.Context, prompt string, modelConfig BedrockModelConfig) ([]byte, error) {
	request := map[string]interface{}{
		"prompt":      prompt,
		"maxTokens":   modelConfig.MaxTokens,
		"temperature": modelConfig.Temperature,
		"topP":        modelConfig.TopP,
	}
	body, err := json.Marshal(request)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}

type AmazonCompletion struct {
	completion ICompletion
}

// Accepts a list of supported model names
func IsModelSupported(modelName string, supportedModels []string) bool {
	for _, supportedModel := range supportedModels {
		if strings.EqualFold(modelName, supportedModel) {
			return true
		}
	}
	return false
}

// Note: The caller should check model support before calling GetCompletion.
func (a *AmazonCompletion) GetCompletion(ctx context.Context, prompt string, modelConfig BedrockModelConfig) ([]byte, error) {
	if a == nil || modelConfig.ModelName == "" {
		return nil, fmt.Errorf("no model name provided to Bedrock completion")
	}
	if strings.Contains(modelConfig.ModelName, "nova") {
		return a.GetNovaCompletion(ctx, prompt, modelConfig)
	} else {
		return a.GetDefaultCompletion(ctx, prompt, modelConfig)
	}
}

func (a *AmazonCompletion) GetDefaultCompletion(ctx context.Context, prompt string, modelConfig BedrockModelConfig) ([]byte, error) {
	request := map[string]interface{}{
		"inputText": fmt.Sprintf("\n\nUser: %s", prompt),
		"textGenerationConfig": map[string]interface{}{
			"maxTokenCount": modelConfig.MaxTokens,
			"temperature":   modelConfig.Temperature,
			"topP":          modelConfig.TopP,
		},
	}
	body, err := json.Marshal(request)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}

func (a *AmazonCompletion) GetNovaCompletion(ctx context.Context, prompt string, modelConfig BedrockModelConfig) ([]byte, error) {
	request := map[string]interface{}{
		"inferenceConfig": map[string]interface{}{
			"max_new_tokens": modelConfig.MaxTokens,
			"temperature":    modelConfig.Temperature,
			"topP":           modelConfig.TopP,
		},
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"text": prompt,
					},
				},
			},
		},
	}
	body, err := json.Marshal(request)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}
