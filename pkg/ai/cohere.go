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
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/cohere-ai/cohere-go"
	"github.com/fatih/color"

	"github.com/k8sgpt-ai/k8sgpt/pkg/cache"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
)

type CohereClient struct {
	client   *cohere.Client
	language string
	model    string
}

func (c *CohereClient) Configure(config IAIConfig, language string) error {
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
	c.language = language
	c.client = client
	c.model = config.GetModel()
	return nil
}

func (c *CohereClient) GetCompletion(ctx context.Context, prompt, promptTmpl string) (string, error) {
	// Create a completion request
	if len(promptTmpl) == 0 {
		promptTmpl = PromptMap["default"]
	}
	resp, err := c.client.Generate(cohere.GenerateOptions{
		Model:             c.model,
		Prompt:            fmt.Sprintf(strings.TrimSpace(promptTmpl), c.language, prompt),
		MaxTokens:         cohere.Uint(2048),
		Temperature:       cohere.Float64(0.75),
		K:                 cohere.Int(0),
		StopSequences:     []string{},
		ReturnLikelihoods: "NONE",
	})
	if err != nil {
		return "", err
	}
	return resp.Generations[0].Text, nil
}

func (a *CohereClient) Parse(ctx context.Context, prompt []string, cache cache.ICache, promptTmpl string) (string, error) {
	inputKey := strings.Join(prompt, " ")
	// Check for cached data
	cacheKey := util.GetCacheKey(a.GetName(), a.language, inputKey)

	if !cache.IsCacheDisabled() && cache.Exists(cacheKey) {
		response, err := cache.Load(cacheKey)
		if err != nil {
			return "", err
		}

		if response != "" {
			output, err := base64.StdEncoding.DecodeString(response)
			if err != nil {
				color.Red("error decoding cached data: %v", err)
				return "", nil
			}
			return string(output), nil
		}
	}

	response, err := a.GetCompletion(ctx, inputKey, promptTmpl)
	if err != nil {
		return "", err
	}

	err = cache.Store(cacheKey, base64.StdEncoding.EncodeToString([]byte(response)))

	if err != nil {
		color.Red("error storing value to cache: %v", err)
		return "", nil
	}

	return response, nil
}

func (a *CohereClient) GetName() string {
	return "cohere"
}
