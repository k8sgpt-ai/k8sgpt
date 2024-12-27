package ai

import (
	"context"
	"errors"

	"github.com/liushuangls/go-anthropic/v2"
	"k8s.io/utils/ptr"
)

const anthropicClientName = "claude"

type ClaudeClient struct {
	client      *anthropic.Client
	model       string
	temperature float32
	topP        float32
	topK        int32
	maxTokens   int
}

func (c *ClaudeClient) Configure(config IAIConfig) error {
	token := config.GetPassword()

	client := anthropic.NewClient(token)
	if client == nil {
		return errors.New("error creating OpenAI client")
	}
	c.client = client
	c.model = config.GetModel()
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()
	c.maxTokens = 2048
	return nil
}

func (c *ClaudeClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	// Create a completion request
	resp, err := c.client.CreateMessages(ctx, anthropic.MessagesRequest{
		Model: anthropic.ModelClaude3Dot5Sonnet20241022,
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage(prompt),
		},
		Temperature: ptr.To(c.temperature),
		TopP:        ptr.To(c.topP),
		TopK:        ptr.To[int](int(c.topK)),
		MaxTokens:   maxToken,
	})
	if err != nil {
		return "", err
	}
	return resp.Content[0].GetText(), nil
}

func (c *ClaudeClient) GetName() string {
	return anthropicClientName
}
