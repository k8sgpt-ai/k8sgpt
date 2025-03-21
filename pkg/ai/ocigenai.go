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
	"fmt"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/generativeai"
	"github.com/oracle/oci-go-sdk/v65/generativeaiinference"
	"reflect"
)

const ociClientName = "oci"

type ociModelVendor string

const (
	vendorCohere = "cohere"
	vendorMeta   = "meta"
)

type OCIGenAIClient struct {
	nopCloser

	client        *generativeaiinference.GenerativeAiInferenceClient
	model         *generativeai.Model
	modelID       string
	compartmentId string
	temperature   float32
	topP          float32
	topK          int32
	maxTokens     int
}

func (c *OCIGenAIClient) GetName() string {
	return ociClientName
}

func (c *OCIGenAIClient) Configure(config IAIConfig) error {
	config.GetEndpointName()
	c.modelID = config.GetModel()
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()
	c.topK = config.GetTopK()
	c.maxTokens = config.GetMaxTokens()
	c.compartmentId = config.GetCompartmentId()
	provider := common.DefaultConfigProvider()
	client, err := generativeaiinference.NewGenerativeAiInferenceClientWithConfigurationProvider(provider)
	if err != nil {
		return err
	}
	c.client = &client
	model, err := c.getModel(provider)
	if err != nil {
		return err
	}
	c.model = model
	return nil
}

func (c *OCIGenAIClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	request := c.newChatRequest(prompt)
	response, err := c.client.Chat(ctx, request)
	if err != nil {
		return "", err
	}
	if err != nil {
		return "", err
	}
	return extractGeneratedText(response.ChatResponse)
}

func (c *OCIGenAIClient) newChatRequest(prompt string) generativeaiinference.ChatRequest {
	return generativeaiinference.ChatRequest{
		ChatDetails: generativeaiinference.ChatDetails{
			CompartmentId: &c.compartmentId,
			ServingMode:   c.getServingMode(),
			ChatRequest:   c.getChatModelRequest(prompt),
		},
	}
}

func (c *OCIGenAIClient) getChatModelRequest(prompt string) generativeaiinference.BaseChatRequest {
	temperatureF64 := float64(c.temperature)
	topPF64 := float64(c.topP)
	topK := int(c.topK)

	switch c.getVendor() {
	case vendorMeta:
		messages := []generativeaiinference.Message{
			generativeaiinference.UserMessage{
				Content: []generativeaiinference.ChatContent{
					generativeaiinference.TextContent{
						Text: &prompt,
					},
				},
			},
		}
		// 0 is invalid for Meta vendor type, instead use -1 to disable topK sampling.
		if topK == 0 {
			topK = -1
		}
		return generativeaiinference.GenericChatRequest{
			Messages:    messages,
			TopK:        &topK,
			TopP:        &topPF64,
			Temperature: &temperatureF64,
			MaxTokens:   &c.maxTokens,
		}
	default: // Default to cohere
		return generativeaiinference.CohereChatRequest{
			Message:     &prompt,
			MaxTokens:   &c.maxTokens,
			Temperature: &temperatureF64,
			TopK:        &topK,
			TopP:        &topPF64,
		}

	}
}

func extractGeneratedText(llmInferenceResponse generativeaiinference.BaseChatResponse) (string, error) {
	switch response := llmInferenceResponse.(type) {
	case generativeaiinference.GenericChatResponse:
		if len(response.Choices) > 0 && len(response.Choices[0].Message.GetContent()) > 0 {
			if content, ok := response.Choices[0].Message.GetContent()[0].(generativeaiinference.TextContent); ok {
				return *content.Text, nil
			}
		}
		return "", errors.New("no text found in oci response")
	case generativeaiinference.CohereChatResponse:
		return *response.Text, nil
	default:
		return "", fmt.Errorf("unknown oci response type: %s", reflect.TypeOf(llmInferenceResponse).Name())
	}
}

func (c *OCIGenAIClient) getServingMode() generativeaiinference.ServingMode {
	if c.isBaseModel() {
		return generativeaiinference.OnDemandServingMode{
			ModelId: &c.modelID,
		}
	}
	return generativeaiinference.DedicatedServingMode{
		EndpointId: &c.modelID,
	}
}

func (c *OCIGenAIClient) getModel(provider common.ConfigurationProvider) (*generativeai.Model, error) {
	client, err := generativeai.NewGenerativeAiClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, err
	}
	response, err := client.GetModel(context.Background(), generativeai.GetModelRequest{
		ModelId: &c.modelID,
	})
	if err != nil {
		return nil, err
	}
	return &response.Model, nil
}

func (c *OCIGenAIClient) isBaseModel() bool {
	return c.model != nil && c.model.Type == generativeai.ModelTypeBase
}

func (c *OCIGenAIClient) getVendor() ociModelVendor {
	if c.model == nil || c.model.Vendor == nil {
		return ""
	}
	return ociModelVendor(*c.model.Vendor)
}
