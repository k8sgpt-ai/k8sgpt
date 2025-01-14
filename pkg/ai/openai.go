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
	// organizationId string
}

const (
	// OpenAI completion parameters
	maxToken         = 2048
	presencePenalty  = 0.0
	frequencyPenalty = 0.0
)

func (c *OpenAIClient) Configure(config IAIConfig) error {
	token := config.GetPassword()
	defaultConfig := openai.DefaultConfig(token)
	orgId := config.GetOrganizationId()
	proxyEndpoint := config.GetProxyEndpoint()

	baseURL := config.GetBaseURL()
	if baseURL != "" {
		defaultConfig.BaseURL = baseURL
	}

	transport := &http.Transport{}
	if proxyEndpoint != "" {
		proxyUrl, err := url.Parse(proxyEndpoint)
		if err != nil {
			return err
		}
		transport.Proxy = http.ProxyURL(proxyUrl)
	}

	if orgId != "" {
		defaultConfig.OrgID = orgId
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
		return errors.New("error creating OpenAI client")
	}
	c.client = client
	c.model = config.GetModel()
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()
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

// OpenAIHeaderTransport is an http.RoundTripper that adds the given headers to each request.
type OpenAIHeaderTransport struct {
	Origin  http.RoundTripper
	Headers []http.Header
}

// RoundTrip implements the http.RoundTripper interface.
func (t *OpenAIHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original request
	clonedReq := req.Clone(req.Context())
	for _, header := range t.Headers {
		for key, values := range header {
			// Possible values per header:  RFC 2616
			for _, value := range values {
				clonedReq.Header.Add(key, value)
			}
		}
	}

	return t.Origin.RoundTrip(clonedReq)
}
