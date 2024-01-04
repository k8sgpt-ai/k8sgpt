package ai

import (
	"context"
	"errors"

	"github.com/sashabaranov/go-openai"
)

type AzureAIClient struct {
	client      *openai.Client
	model       string
	temperature float32
}

func (c *AzureAIClient) Configure(config IAIConfig) error {
	token := config.GetPassword()
	baseURL := config.GetBaseURL()
	engine := config.GetEngine()
	defaultConfig := openai.DefaultAzureConfig(token, baseURL)

	defaultConfig.AzureModelMapperFunc = func(model string) string {
		// If you use a deployment name different from the model name, you can customize the AzureModelMapperFunc function
		azureModelMapping := map[string]string{
			model: engine,
		}
		return azureModelMapping[model]

	}
	client := openai.NewClientWithConfig(defaultConfig)
	if client == nil {
		return errors.New("error creating Azure OpenAI client")
	}
	c.client = client
	c.model = config.GetModel()
	c.temperature = config.GetTemperature()
	return nil
}

func (c *AzureAIClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	// Create a completion request
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: c.temperature,
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func (c *AzureAIClient) GetName() string {
	return "azureopenai"
}
