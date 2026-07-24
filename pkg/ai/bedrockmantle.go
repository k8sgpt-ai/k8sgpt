package ai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/sashabaranov/go-openai"
)

const bedrockMantleClientName = "bedrockmantle"

type AmazonBedrockMantleClient struct {
	nopCloser

	client      *openai.Client
	model       string
	temperature float32
	topP        float32
}

func (c *AmazonBedrockMantleClient) Configure(config IAIConfig) error {
	region := getRegion(config.GetProviderRegion())
	if region == "" {
		return errors.New("provider region is required for bedrockmantle (e.g. us-east-1)")
	}

	token := os.Getenv("AWS_BEARER_TOKEN_BEDROCK")
	if token == "" {
		return errors.New("AWS_BEARER_TOKEN_BEDROCK environment variable is required for bedrockmantle")
	}
	defaultConfig := openai.DefaultConfig(token)

	baseURL := config.GetBaseURL()
	if baseURL != "" {
		defaultConfig.BaseURL = baseURL
	} else {
		defaultConfig.BaseURL = fmt.Sprintf("https://bedrock-mantle.%s.api.aws/v1", region)
	}

	transport := &http.Transport{}
	if proxyEndpoint := config.GetProxyEndpoint(); proxyEndpoint != "" {
		proxyUrl, err := url.Parse(proxyEndpoint)
		if err != nil {
			return err
		}
		transport.Proxy = http.ProxyURL(proxyUrl)
	}

	defaultConfig.HTTPClient = &http.Client{
		Transport: &OpenAIHeaderTransport{
			Origin:  transport,
			Headers: config.GetCustomHeaders(),
		},
	}

	client := openai.NewClientWithConfig(defaultConfig)
	if client == nil {
		return errors.New("error creating Bedrock Mantle client")
	}
	c.client = client
	c.model = config.GetModel()
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()
	return nil
}

func (c *AmazonBedrockMantleClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature:      c.temperature,
		MaxTokens:        maxToken,
		PresencePenalty:  presencePenalty,
		FrequencyPenalty: frequencyPenalty,
		TopP:             c.topP,
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func (c *AmazonBedrockMantleClient) GetName() string {
	return bedrockMantleClientName
}
