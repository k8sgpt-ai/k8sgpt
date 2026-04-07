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

const groqAIClientName = "groq"

// Default Groq API endpoint (OpenAI-compatible)
const groqAPIBaseURL = "https://api.groq.com/openai/v1"

type GroqClient struct {
	nopCloser

	client      *openai.Client
	model       string
	temperature float32
	topP        float32
}

func (c *GroqClient) Configure(config IAIConfig) error {
	token := config.GetPassword()
	defaultConfig := openai.DefaultConfig(token)
	proxyEndpoint := config.GetProxyEndpoint()

	baseURL := config.GetBaseURL()
	if baseURL != "" {
		defaultConfig.BaseURL = baseURL
	} else {
		defaultConfig.BaseURL = groqAPIBaseURL
	}

	transport := &http.Transport{}
	if proxyEndpoint != "" {
		proxyUrl, err := url.Parse(proxyEndpoint)
		if err != nil {
			return err
		}
		transport.Proxy = http.ProxyURL(proxyUrl)
	}

	customHeaders := config.GetCustomHeaders()
	defaultConfig.HTTPClient = &http.Client{
		Transport: &OpenAIHeaderTransport{
			Origin:  transport,
			Headers: customHeaders,
		},
	}

	client := openai.NewClientWithConfig(defaultConfig)
	if client == nil {
		return errors.New("error creating Groq client")
	}
	c.client = client
	c.model = config.GetModel()
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()
	return nil
}

func (c *GroqClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
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

func (c *GroqClient) GetName() string {
	return groqAIClientName
}
