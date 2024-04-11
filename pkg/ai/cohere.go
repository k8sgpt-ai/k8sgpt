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

	api "github.com/cohere-ai/cohere-go/v2"
	cohere "github.com/cohere-ai/cohere-go/v2/client"
	"github.com/cohere-ai/cohere-go/v2/option"
)

const cohereAIClientName = "cohere"

type CohereClient struct {
	nopCloser

	client      *cohere.Client
	model       string
	temperature float32
	maxTokens   int
}

func (c *CohereClient) Configure(config IAIConfig) error {
	token := config.GetPassword()

	opts := []option.RequestOption{
		cohere.WithToken(token),
	}

	baseURL := config.GetBaseURL()
	if baseURL != "" {
		opts = append(opts, cohere.WithBaseURL(baseURL))
	}

	client := cohere.NewClient(opts...)
	if client == nil {
		return errors.New("error creating Cohere client")
	}

	c.client = client
	c.model = config.GetModel()
	c.temperature = config.GetTemperature()
	c.maxTokens = config.GetMaxTokens()

	return nil
}

func (c *CohereClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	// Create a completion request
	response, err := c.client.Chat(ctx, &api.ChatRequest{
		Message:      prompt,
		Model:        &c.model,
		K:            api.Int(0),
		Preamble:     api.String(""),
		Temperature:  api.Float64(float64(c.temperature)),
		RawPrompting: api.Bool(false),
		MaxTokens:    api.Int(c.maxTokens),
	})
	if err != nil {
		return "", err
	}
	return response.Text, nil
}

func (c *CohereClient) GetName() string {
	return cohereAIClientName
}
