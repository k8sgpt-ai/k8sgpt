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

	"github.com/cohere-ai/cohere-go"
)

const cohereAIClientName = "cohere"

type CohereClient struct {
	nopCloser

	client      *cohere.Client
	model       string
	temperature float32
}

func (c *CohereClient) Configure(config IAIConfig) error {
	token := config.GetPassword()

	client, err := cohere.CreateClient(token)
	if err != nil {
		return err
	}

	baseURL := config.GetBaseURL()
	if baseURL != "" {
		client.BaseURL = baseURL
	}

	if client == nil {
		return errors.New("error creating Cohere client")
	}
	c.client = client
	c.model = config.GetModel()
	c.temperature = config.GetTemperature()
	return nil
}

func (c *CohereClient) GetCompletion(_ context.Context, prompt string) (string, error) {
	// Create a completion request
	resp, err := c.client.Generate(cohere.GenerateOptions{
		Model:             c.model,
		Prompt:            prompt,
		MaxTokens:         cohere.Uint(2048),
		Temperature:       cohere.Float64(float64(c.temperature)),
		K:                 cohere.Int(0),
		StopSequences:     []string{},
		ReturnLikelihoods: "NONE",
	})
	if err != nil {
		return "", err
	}
	return resp.Generations[0].Text, nil
}

func (c *CohereClient) GetName() string {
	return cohereAIClientName
}
