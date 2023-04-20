package ai

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/k8sgpt-ai/k8sgpt/pkg/util"

	"github.com/fatih/color"
	"github.com/spf13/viper"

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

func (c *AzureAIClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
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

func (a *AzureAIClient) Parse(ctx context.Context, prompt []string, nocache bool) (string, error) {
	inputKey := strings.Join(prompt, " ")
	// Check for cached data
	sEnc := base64.StdEncoding.EncodeToString([]byte(inputKey))
	cacheKey := util.GetCacheKey(a.GetName(), a.language, sEnc)
	// find in viper cache
	if viper.IsSet(cacheKey) && !nocache {
		// retrieve data from cache
		response := viper.GetString(cacheKey)
		if response == "" {
			color.Red("error retrieving cached data")
			return "", nil
		}
		output, err := base64.StdEncoding.DecodeString(response)
		if err != nil {
			color.Red("error decoding cached data: %v", err)
			return "", nil
		}
		return string(output), nil
	}

	response, err := a.GetCompletion(ctx, inputKey)
	if err != nil {
		return "", err
	}

	if !viper.IsSet(cacheKey) || nocache {
		viper.Set(cacheKey, base64.StdEncoding.EncodeToString([]byte(response)))
		if err := viper.WriteConfig(); err != nil {
			color.Red("error writing config: %v", err)
			return "", nil
		}
	}
	return response, nil
}

func (a *AzureAIClient) GetName() string {
	return "azureopenai"
}
