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
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/generativeaiinference"
	"strings"
)

const ociClientName = "oci"

type OCIGenAIClient struct {
	nopCloser

	client        *generativeaiinference.GenerativeAiInferenceClient
	model         string
	compartmentId string
	temperature   float32
	topP          float32
	maxTokens     int
}

func (c *OCIGenAIClient) GetName() string {
	return ociClientName
}

func (c *OCIGenAIClient) Configure(config IAIConfig) error {
	config.GetEndpointName()
	c.model = config.GetModel()
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()
	c.maxTokens = config.GetMaxTokens()
	c.compartmentId = config.GetCompartmentId()
	provider := common.DefaultConfigProvider()
	client, err := generativeaiinference.NewGenerativeAiInferenceClientWithConfigurationProvider(provider)
	if err != nil {
		return err
	}
	c.client = &client
	return nil
}

func (c *OCIGenAIClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	generateTextRequest := c.newGenerateTextRequest(prompt)
	generateTextResponse, err := c.client.GenerateText(ctx, generateTextRequest)
	if err != nil {
		return "", err
	}
	return extractGeneratedText(generateTextResponse.InferenceResponse)
}

func (c *OCIGenAIClient) newGenerateTextRequest(prompt string) generativeaiinference.GenerateTextRequest {
	temperatureF64 := float64(c.temperature)
	topPF64 := float64(c.topP)
	return generativeaiinference.GenerateTextRequest{
		GenerateTextDetails: generativeaiinference.GenerateTextDetails{
			CompartmentId: &c.compartmentId,
			ServingMode: generativeaiinference.OnDemandServingMode{
				ModelId: &c.model,
			},
			InferenceRequest: generativeaiinference.CohereLlmInferenceRequest{
				Prompt:      &prompt,
				MaxTokens:   &c.maxTokens,
				Temperature: &temperatureF64,
				TopP:        &topPF64,
			},
		},
	}
}

func extractGeneratedText(llmInferenceResponse generativeaiinference.LlmInferenceResponse) (string, error) {
	response, ok := llmInferenceResponse.(generativeaiinference.CohereLlmInferenceResponse)
	if !ok {
		return "", errors.New("failed to extract generated text from backed response")
	}
	sb := strings.Builder{}
	for _, text := range response.GeneratedTexts {
		if text.Text != nil {
			sb.WriteString(*text.Text)
		}
	}
	return sb.String(), nil
}
