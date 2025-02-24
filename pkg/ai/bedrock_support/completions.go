package bedrock_support

import (
	"context"
	"encoding/json"
	"fmt"
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

func (a *AmazonCompletion) GetCompletion(ctx context.Context, prompt string, modelConfig BedrockModelConfig) ([]byte, error) {
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
