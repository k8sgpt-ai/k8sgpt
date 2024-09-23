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

	//Due to the temporary absence of a Golang client, the OpenAI client is being used for compatibility.
	"github.com/sashabaranov/go-openai"
)

const alicloudAIClientName = "alicloud"

type DashScopeClient struct {
	nopCloser

	client     *openai.Client
	model      string
	temprature float32
	topP       float32
	topK       int32
}

func (c *DashScopeClient) Configure(config IAIConfig) error {
	const defaultBaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"

	apiKey := config.GetPassword()
	defaultConfig := openai.DefaultConfig(apiKey)
	defaultConfig.BaseURL = defaultBaseURL

	baseURL := config.GetBaseURL()
	if baseURL != "" {
		defaultConfig.BaseURL = baseURL
	}

	client := openai.NewClientWithConfig(defaultConfig)
	if client == nil {
		return errors.New("error create DashScope client")
	}
	c.client = client
	c.temprature = config.GetTemperature()
	c.topP = config.GetTopP()
	c.model = config.GetModel()

	return nil
}

func (c *DashScopeClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: c.temprature,
		TopP:        c.topP,
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func (c *DashScopeClient) GetName() string {
	return alicloudAIClientName
}
