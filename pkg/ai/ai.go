package ai

import (
	"context"
	"errors"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

const (
	default_prompt = "Simplify the following Kubernetes error message and provide a solution in %s: %s"
	prompt_a       = "Read the following input %s and provide possible scenarios for remediation in %s"
	prompt_b       = "Considering the following input from the Kubernetes resource %s and the error message %s, provide possible scenarios for remediation in %s"
	prompt_c       = "Reading the following %s error message and it's accompanying log message %s, how would you simplify this message?"
)

type AIConfiguration struct {
	Providers []AIProvider `mapstructure:"providers"`
}

type AIProvider struct {
	Name     string `mapstructure:"name"`
	Model    string `mapstructure:"model"`
	Password string `mapstructure:"password"`
}

type OpenAIClient struct {
	client   *openai.Client
	language string
	model    string
}

func (c *OpenAIClient) Configure(token string, model string, language string) error {
	client := openai.NewClient(token)
	if client == nil {
		return errors.New("error creating OpenAI client")
	}
	c.language = language
	c.client = client
	c.model = model
	return nil
}

func (c *OpenAIClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	// Create a completion request
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: fmt.Sprintf(default_prompt, c.language, prompt),
			},
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
