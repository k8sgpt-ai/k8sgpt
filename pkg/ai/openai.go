/*
Copyright 2023 The K8sGPT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ai

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/sashabaranov/go-openai"
)

const openAIClientName = "openai"

type OpenAIClient struct {
	nopCloser

	client      *openai.Client
	model       string
	temperature float32
	topP        float32
}

const (
	// OpenAI completion parameters
	maxToken         = 2048
	presencePenalty  = 0.0
	frequencyPenalty = 0.0
)

func (c *OpenAIClient) Configure(config IAIConfig, index int) error {
	token := config.GetPassword(index)
	defaultConfig := openai.DefaultConfig(token)
	proxyEndpoint := config.GetProxyEndpoint(index)

	baseURL := config.GetBaseURL(index)
	if baseURL != "" {
		defaultConfig.BaseURL = baseURL
	}

	if proxyEndpoint != "" {
		proxyUrl, err := url.Parse(proxyEndpoint)
		if err != nil {
			return err
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}

		defaultConfig.HTTPClient = &http.Client{
			Transport: transport,
		}
	}

	client := openai.NewClientWithConfig(defaultConfig)
	if client == nil {
		return errors.New("error creating OpenAI client")
	}
	c.client = client
	c.model = config.GetModel(index)
	c.temperature = config.GetTemperature(index)
	c.topP = config.GetTopP(index)
	return nil
}

func (c *OpenAIClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	// Create a completion request
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

func (c *OpenAIClient) GetName() string {
	return openAIClientName
}
