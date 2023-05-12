package ai

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/k8sgpt-ai/k8sgpt/pkg/cache"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"

	"github.com/fatih/color"

	"github.com/sashabaranov/go-openai"
)

type AzureAIClient struct {
	client   *openai.Client
	language string
	model    string
}

func (c *AzureAIClient) Configure(config IAIConfig, lang string) error {
	token := config.GetPassword()
	baseURL := config.GetBaseURL()
	engine := config.GetEngine()
	defaultConfig := openai.DefaultAzureConfig(token, baseURL, engine)
	client := openai.NewClientWithConfig(defaultConfig)
	if client == nil {
		return errors.New("error creating Azure OpenAI client")
	}
	c.language = lang
	c.client = client
	c.model = config.GetModel()
	return nil
}

func (c *AzureAIClient) GetCompletion(ctx context.Context, prompt string, promptTmpl string) (string, error) {
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

func (a *AzureAIClient) Parse(ctx context.Context, prompt []string, cache cache.ICache, promptTmpl string) (string, error) {
	inputKey := strings.Join(prompt, " ")
	// Check for cached data
	cacheKey := util.GetCacheKey(a.GetName(), a.language, inputKey)

	if !cache.IsCacheDisabled() && cache.Exists(cacheKey) {
		response, err := cache.Load(cacheKey)
		if err != nil {
			return "", err
		}

		if response != "" {
			output, err := base64.StdEncoding.DecodeString(response)
			if err != nil {
				color.Red("error decoding cached data: %v", err)
				return "", nil
			}
			return string(output), nil
		}
	}

	response, err := a.GetCompletion(ctx, inputKey, promptTmpl)
	if err != nil {
		return "", err
	}

	err = cache.Store(cacheKey, base64.StdEncoding.EncodeToString([]byte(response)))

	if err != nil {
		color.Red("error storing value to cache: %v", err)
		return "", nil
	}

	return response, nil
}

func (a *AzureAIClient) GetName() string {
	return "azureopenai"
}
