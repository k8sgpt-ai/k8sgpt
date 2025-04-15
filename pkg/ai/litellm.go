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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const liteLLMClientName = "litellm"

type LiteLLMClient struct {
	nopCloser

	client      *http.Client
	baseURL     string
	model       string
	apiKey      string
	temperature float32
	topP        float32
}

const (
	defaultLiteLLMBaseURL = "http://localhost:4000"
	defaultLiteLLMModel   = "gpt-3.5-turbo" // Default model, can be configured
)

type LiteLLMRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float32   `json:"temperature,omitempty"`
	TopP        float32   `json:"top_p,omitempty"`
}

type LiteLLMResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *LiteLLMClient) Configure(config IAIConfig) error {
	baseURL := config.GetBaseURL()
	if baseURL == "" {
		baseURL = defaultLiteLLMBaseURL
	}
	_, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	c.baseURL = baseURL

	proxyEndpoint := config.GetProxyEndpoint()
	httpClient := http.DefaultClient
	if proxyEndpoint != "" {
		proxyUrl, err := url.Parse(proxyEndpoint)
		if err != nil {
			return err
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
		httpClient = &http.Client{
			Transport: transport,
		}
	}

	c.client = httpClient
	c.model = config.GetModel()
	if c.model == "" {
		c.model = defaultLiteLLMModel
	}
	c.apiKey = config.GetPassword()
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()
	return nil
}

func (c *LiteLLMClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	reqBody := LiteLLMRequest{
		Model: c.model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: c.temperature,
		TopP:        c.topP,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/v1/chat/completions", c.baseURL), bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LiteLLM API error: %s (status code: %d)", string(body), resp.StatusCode)
	}

	var llmResp LiteLLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return "", err
	}

	if len(llmResp.Choices) == 0 {
		return "", errors.New("no completion choices returned from LiteLLM")
	}

	return llmResp.Choices[0].Message.Content, nil
}

func (c *LiteLLMClient) GetName() string {
	return liteLLMClientName
}
