package bedrock_support

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

var SUPPPORTED_BEDROCK_MODELS = []string{
	"anthropic.claude-3-5-sonnet-20240620-v1:0",
	"us.anthropic.claude-3-5-sonnet-20241022-v2:0",
	"anthropic.claude-v2",
	"anthropic.claude-v1",
	"anthropic.claude-instant-v1",
	"ai21.j2-ultra-v1",
	"ai21.j2-jumbo-instruct",
	"amazon.titan-text-express-v1",
	"amazon.nova-pro-v1:0",
	"eu.amazon.nova-lite-v1:0",
}

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

func isModelSupported(modelName string) bool {
	for _, supportedModel := range SUPPPORTED_BEDROCK_MODELS {
		if modelName == supportedModel {
			return true
		}
	}
	return false
}

func (a *AmazonCompletion) GetCompletion(ctx context.Context, prompt string, modelConfig BedrockModelConfig) ([]byte, error) {
	if !isModelSupported(modelConfig.ModelName) {
		return nil, fmt.Errorf("model %s is not supported", modelConfig.ModelName)
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
