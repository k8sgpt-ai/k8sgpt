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
	"strings"
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
	modelId       string
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
	c.modelId = config.GetModel()
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
	generateTextRequest := c.newGenerateTextRequest(prompt)
	generateTextResponse, err := c.client.GenerateText(ctx, generateTextRequest)
	if err != nil {
		return "", err
	}
	return c.extractGeneratedText(generateTextResponse.InferenceResponse)
}

func (c *OCIGenAIClient) newGenerateTextRequest(prompt string) generativeaiinference.GenerateTextRequest {
	return generativeaiinference.GenerateTextRequest{
		GenerateTextDetails: generativeaiinference.GenerateTextDetails{
			CompartmentId:    &c.compartmentId,
			ServingMode:      c.getServingMode(),
			InferenceRequest: c.getInferenceRequest(prompt),
		},
	}
}

func (c *OCIGenAIClient) getServingMode() generativeaiinference.ServingMode {
	if c.isBaseModel() {
		return generativeaiinference.OnDemandServingMode{
			ModelId: &c.modelId,
		}
	}
	return generativeaiinference.DedicatedServingMode{
		EndpointId: &c.modelId,
	}
}

func (c *OCIGenAIClient) getInferenceRequest(prompt string) generativeaiinference.LlmInferenceRequest {
	temperatureF64 := float64(c.temperature)
	topPF64 := float64(c.topP)
	topK := int(c.topP)

	switch c.getVendor() {
	case vendorMeta:
		return generativeaiinference.LlamaLlmInferenceRequest{
			Prompt:      &prompt,
			MaxTokens:   &c.maxTokens,
			Temperature: &temperatureF64,
			TopK:        &topK,
			TopP:        &topPF64,
		}
	default: // Default to cohere
		return generativeaiinference.CohereLlmInferenceRequest{
			Prompt:      &prompt,
			MaxTokens:   &c.maxTokens,
			Temperature: &temperatureF64,
			TopK:        &topK,
			TopP:        &topPF64,
		}
	}
}

func (c *OCIGenAIClient) getModel(provider common.ConfigurationProvider) (*generativeai.Model, error) {
	client, err := generativeai.NewGenerativeAiClientWithConfigurationProvider(provider)
	if err != nil {
		return nil, err
	}
	response, err := client.GetModel(context.Background(), generativeai.GetModelRequest{
		ModelId: &c.modelId,
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

func (c *OCIGenAIClient) extractGeneratedText(llmInferenceResponse generativeaiinference.LlmInferenceResponse) (string, error) {
	switch response := llmInferenceResponse.(type) {
	case generativeaiinference.LlamaLlmInferenceResponse:
		if len(response.Choices) > 0 && response.Choices[0].Text != nil {
			return *response.Choices[0].Text, nil
		}
		return "", errors.New("no text found in oci response")
	case generativeaiinference.CohereLlmInferenceResponse:
		sb := strings.Builder{}
		for _, text := range response.GeneratedTexts {
			if text.Text != nil {
				sb.WriteString(*text.Text)
			}
		}
		return sb.String(), nil
	default:
		return "", fmt.Errorf("unknown oci response type: %s", reflect.TypeOf(llmInferenceResponse).Name())
	}
}
