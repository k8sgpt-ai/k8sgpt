package ai

import (
	"context"
	"log"
	"time"

	qwen "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const (
	qwenAIClientName     = "qwenai"
	qwenMaxToken         = maxToken
	qwenPresencePenalty  = 0.0
	qwenFrequencyPenalty = 0.0
	qwenTimeout          = 120
)

type QwenAIClient struct {
	nopCloser
	client         qwen.Client
	model          string
	temperature    float32
	topP           float32
	enableThinking bool
	tools          []qwen.ChatCompletionToolUnionParam
}

func (c *QwenAIClient) Configure(config IAIConfig) (err error) {
	c.client = qwen.NewClient(
		option.WithAPIKey(config.GetPassword()),
		option.WithBaseURL(config.GetBaseURL()),
		option.WithRequestTimeout(qwenTimeout*time.Second),
	)
	c.model = config.GetModel()
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()
	return nil
}

func (c *QwenAIClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	params := qwen.ChatCompletionNewParams{
		Messages: []qwen.ChatCompletionMessageParamUnion{
			qwen.UserMessage(prompt),
		},
		Seed:                qwen.Int(0),
		Model:               c.model,
		Temperature:         qwen.Float(float64(c.temperature)),
		MaxTokens:           qwen.Int(int64(qwenMaxToken)),
		TopP:                qwen.Float(float64(c.topP)),
		MaxCompletionTokens: qwen.Int(int64(qwenMaxToken)),
		FrequencyPenalty:    qwen.Float(float64(qwenFrequencyPenalty)),
		PresencePenalty:     qwen.Float(float64(qwenPresencePenalty)),
	}
	completion, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		log.Fatalf("Qwen failed to complete, %v", err)
		return "", err
	}

	return completion.Choices[0].Message.Content, nil
}

func (c *QwenAIClient) GetName() string {
	return qwenAIClientName
}
