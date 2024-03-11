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
	"fmt"

	"cloud.google.com/go/vertexai/genai"
	"github.com/fatih/color"
)

const googleVertexAIClientName = "googlevertexai"

type GoogleVertexAIClient struct {
	client *genai.Client

	model       string
	temperature float32
	topP        float32
	maxTokens   int
}

// Vertex AI Gemini supported Regions
// https://cloud.google.com/vertex-ai/docs/generative-ai/model-reference/gemini
const VERTEXAI_DEFAULT_REGION = "us-central1" // default use us-east-1 region

const (
	US_Central_1             = "us-central1"
	US_West_4                = "us-west4"
	North_America_Northeast1 = "northamerica-northeast1"
	US_East_4                = "us-east4"
	US_West_1                = "us-west1"
	Asia_Northeast_3         = "asia-northeast3"
	Asia_Southeast_1         = "asia-southeast1"
	Asia_Northeast_1         = "asia-northeast1"
)

var VERTEXAI_SUPPORTED_REGION = []string{
	US_Central_1,
	US_West_4,
	North_America_Northeast1,
	US_East_4,
	US_West_1,
	Asia_Northeast_3,
	Asia_Southeast_1,
	Asia_Northeast_1,
}

const (
	ModelGeminiProV1 = "gemini-1.0-pro-001"
)

var VERTEXAI_MODELS = []string{
	ModelGeminiProV1,
}

// GetModelOrDefault check config model
func GetVertexAIModelOrDefault(model string) string {

	// Check if the provided model is in the list
	for _, m := range VERTEXAI_MODELS {
		if m == model {
			return model // Return the provided model
		}
	}

	// Return the default model if the provided model is not in the list
	return VERTEXAI_MODELS[0]
}

// GetModelOrDefault check config region
func GetVertexAIRegionOrDefault(region string) string {

	// Check if the provided model is in the list
	for _, m := range VERTEXAI_SUPPORTED_REGION {
		if m == region {
			return region // Return the provided model
		}
	}

	// Return the default model if the provided model is not in the list
	return VERTEXAI_DEFAULT_REGION
}

func (g *GoogleVertexAIClient) Configure(config IAIConfig) error {
	ctx := context.Background()

	// Currently you can access VertexAI either by being authenticated via OAuth or Bearer token so we need to consider both
	projectId := config.GetProviderId()
	region := GetVertexAIRegionOrDefault(config.GetProviderRegion())

	client, err := genai.NewClient(ctx, projectId, region)
	if err != nil {
		return fmt.Errorf("creating genai Google SDK client: %w", err)
	}

	g.client = client
	g.model = GetVertexAIModelOrDefault(config.GetModel())
	g.temperature = config.GetTemperature()
	g.topP = config.GetTopP()
	g.maxTokens = config.GetMaxTokens()

	return nil
}

func (g *GoogleVertexAIClient) GetCompletion(ctx context.Context, prompt string) (string, error) {

	model := g.client.GenerativeModel(g.model)
	model.SetTemperature(g.temperature)
	model.SetTopP(g.topP)
	model.SetMaxOutputTokens(int32(g.maxTokens))

	// Google AI SDK is capable of different inputs than just text, for now set explicit text prompt type.
	// Similarly, we could stream the response. For now k8sgpt does not support streaming.
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 {
		if resp.PromptFeedback.BlockReason > 0 {
			for _, r := range resp.PromptFeedback.SafetyRatings {
				if !r.Blocked {
					continue
				}
				return "", fmt.Errorf("complection blocked due to %v with probability %v", r.Category.String(), r.Probability.String())
			}
		}
		return "", errors.New("no complection returned; unknown reason")
	}

	// Format output.
	// TODO(bwplotka): Provider richer output in certain cases e.g. suddenly finished
	// completion based on finish reasons or safety rankings.
	got := resp.Candidates[0]
	var output string
	for _, part := range got.Content.Parts {
		switch o := part.(type) {
		case genai.Text:
			output += string(o)
			output += "\n"
		default:
			color.Yellow("found unsupported AI response part of type %T; ignoring", part)
		}
	}

	if got.CitationMetadata != nil && len(got.CitationMetadata.Citations) > 0 {
		output += "Citations:\n"
		for _, source := range got.CitationMetadata.Citations {
			// TODO(bwplotka): Give details around what exactly words could be attributed to the citation.
			output += fmt.Sprintf("* %s, %s\n", source.URI, source.License)
		}
	}
	return output, nil
}

func (g *GoogleVertexAIClient) GetName() string {
	return googleVertexAIClientName
}

func (g *GoogleVertexAIClient) Close() {
	if err := g.client.Close(); err != nil {
		color.Red("googleai client close error: %v", err)
	}
}
