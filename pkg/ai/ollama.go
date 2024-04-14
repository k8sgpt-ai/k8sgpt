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

	ollama "github.com/ollama/ollama/api"
)

const ollamaClientName = "ollama"

type OllamaClient struct {
	nopCloser

	client      *ollama.Client
	model       string
	temperature float32
	topP        float32
}

const (
	defaultBaseURL = "http://localhost:11434"
	defaultModel   = "llama3"
)

func (c *OllamaClient) Configure(config IAIConfig) error {
	baseURL := config.GetBaseURL()
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	baseClientURL, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

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

	c.client = ollama.NewClient(baseClientURL, httpClient)
	if c.client == nil {
		return errors.New("error creating Ollama client")
	}
	c.model = config.GetModel()
	if c.model == "" {
		c.model = defaultModel
	}
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()
	return nil
}
func (c *OllamaClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	req := &ollama.GenerateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: new(bool),
		Options: map[string]interface{}{
			"temperature": c.temperature,
			"top_p":       c.topP,
		},
	}
	completion := ""
	respFunc := func(resp ollama.GenerateResponse) error {
		completion = resp.Response
		return nil
	}
	err := c.client.Generate(ctx, req, respFunc)
	if err != nil {
		return "", err
	}
	return completion, nil
}
func (a *OllamaClient) GetName() string {
	return ollamaClientName
}
