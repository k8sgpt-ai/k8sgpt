package openai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

type Client struct {
	client *openai.Client
}

func (c *Client) GetClient() *openai.Client {
	return c.client
}

func NewClient() (*Client, error) {

	// get the token with viper
	token := viper.GetString("openai_api_key")
	// check if nil
	if token == "" {
		return nil, fmt.Errorf("no OpenAI API Key found")
	}

	client := openai.NewClient(token)
	return &Client{
		client: client,
	}, nil
}

func (c *Client) GetCompletion(ctx context.Context, prompt string) (string, error) {
	// Create a completion request
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: "Simplify the following Kubernetes error message and provide a solution: " + prompt,
			},
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
