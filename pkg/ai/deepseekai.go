package ai

import (
	"context"
	"time"

	deepseek "github.com/cohesion-org/deepseek-go"
)

const deepSeekAIClientName = "deepseekai"

type DeepseekAIClient struct {
	nopCloser

	client      *deepseek.Client
	model       string
	temperature float32
	topP        float32
}

const (
	dsMaxToken         = 2048
	dsPresencePenalty  = 0.0
	dsFrequencyPenalty = 0.0
	dsTimeout          = 120
)

func (c *DeepseekAIClient) Configure(config IAIConfig) (err error) {
	token := config.GetPassword()
	baseURL := config.GetBaseURL()
	dsTimeout := deepseek.WithTimeout(dsTimeout * time.Second)
	if baseURL != "" {
		c.client, err = deepseek.NewClientWithOptions(token,
			deepseek.WithBaseURL(baseURL),
			dsTimeout,
		)
		if err != nil {
			return err
		}
	} else {
		c.client, err = deepseek.NewClientWithOptions(token, dsTimeout)
		if err != nil {
			return err
		}
	}

	c.model = config.GetModel()
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()
	return nil
}

func (c *DeepseekAIClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	request := &deepseek.ChatCompletionRequest{
		Model: c.model,
		Messages: []deepseek.ChatCompletionMessage{
			{Role: deepseek.ChatMessageRoleUser, Content: prompt},
		},
		TopP:             c.topP,
		Temperature:      c.temperature,
		MaxTokens:        dsMaxToken,
		PresencePenalty:  dsPresencePenalty,
		FrequencyPenalty: dsFrequencyPenalty,
	}

	response, err := c.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", err
	}

	return response.Choices[0].Message.Content, nil
}

func (c *DeepseekAIClient) GetName() string {
	return deepSeekAIClientName
}
