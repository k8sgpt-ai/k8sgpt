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

type OpenAIClient struct {
	client   *openai.Client
	language string
	model    string
}

func (c *OpenAIClient) Configure(config IAIConfig, language string) error {
	token := config.GetPassword()
	defaultConfig := openai.DefaultConfig(token)

	baseURL := config.GetBaseURL()
	if baseURL != "" {
		defaultConfig.BaseURL = baseURL
	}

	client := openai.NewClientWithConfig(defaultConfig)
	if client == nil {
		return errors.New("error creating OpenAI client")
	}
	c.language = language
	c.client = client
	c.model = config.GetModel()
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

func (a *OpenAIClient) Parse(ctx context.Context, prompt []string, nocache bool) (string, error) {
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

func (a *OpenAIClient) GetName() string {
	return "openai"
}
