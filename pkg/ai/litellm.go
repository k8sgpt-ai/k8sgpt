/*
Copyright 2024 The K8sGPT Authors.
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

const liteLLMClientName = "litellm"

// Default LiteLLM proxy endpoint (OpenAI-compatible). The LiteLLM proxy exposes
// an OpenAI-compatible API, so it is reached through the same go-openai client
// as the OpenAI backend, only with a different base URL. Point `baseurl` at your
// proxy to route k8sgpt through any of the 100+ providers LiteLLM supports.
const liteLLMBaseURL = "http://localhost:4000/v1"

type LiteLLMClient struct {
	nopCloser

	client      *openai.Client
	model       string
	temperature float32
	topP        float32
}

func (c *LiteLLMClient) Configure(config IAIConfig) error {
	token := config.GetPassword()
	defaultConfig := openai.DefaultConfig(token)
	proxyEndpoint := config.GetProxyEndpoint()

	baseURL := config.GetBaseURL()
	if baseURL != "" {
		defaultConfig.BaseURL = baseURL
	} else {
		defaultConfig.BaseURL = liteLLMBaseURL
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
		return errors.New("error creating LiteLLM client")
	}
	c.client = client
	c.model = config.GetModel()
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()
	return nil
}

func (c *LiteLLMClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
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
	if len(resp.Choices) == 0 {
		return "", errors.New("no completion choices returned from LiteLLM")
	}
	return resp.Choices[0].Message.Content, nil
}

func (c *LiteLLMClient) GetName() string {
	return liteLLMClientName
}
