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
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/cache"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
)

type NoOpAIClient struct {
	client   string
	language string
	model    string
}

func (c *NoOpAIClient) Configure(config IAIConfig, language string) error {
	token := config.GetPassword()
	c.language = language
	c.client = fmt.Sprintf("I am a noop client with the token %s ", token)
	c.model = config.GetModel()
	return nil
}

func (c *NoOpAIClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	// Create a completion request
	response := "I am a noop response to the prompt " + prompt
	return response, nil
}

func (a *NoOpAIClient) Parse(ctx context.Context, prompt []string, cache cache.ICache) (string, error) {
	// parse the text with the AI backend
	inputKey := strings.Join(prompt, " ")
	// Check for cached data
	sEnc := base64.StdEncoding.EncodeToString([]byte(inputKey))
	cacheKey := util.GetCacheKey(a.GetName(), a.language, sEnc)

	response, err := a.GetCompletion(ctx, inputKey)
	if err != nil {
		color.Red("error getting completion: %v", err)
		return "", err
	}

  err = cache.Store(cacheKey, base64.StdEncoding.EncodeToString([]byte(response)))

	if err != nil {
		color.Red("error storing value to cache: %v", err)
		return "", nil
	}

	return response, nil
}

func (a *NoOpAIClient) GetName() string {
	return "noopai"
}
