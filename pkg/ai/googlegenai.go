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

	"github.com/fatih/color"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const googleAIClientName = "google"

type GoogleGenAIClient struct {
	client *genai.Client

	model       string
	temperature float32
	topP        float32
	maxTokens   int
}

func (c *GoogleGenAIClient) Configure(config IAIConfig) error {
	ctx := context.Background()

	// Access your API key as an environment variable (see "Set up your API key" above)
	token := config.GetPassword()
	authOption := option.WithAPIKey(token)
	if token[0] == '{' {
		authOption = option.WithCredentialsJSON([]byte(token))
	}

	client, err := genai.NewClient(ctx, authOption)
	if err != nil {
		return fmt.Errorf("creating genai Google SDK client: %w", err)
	}

	c.client = client
	c.model = config.GetModel()
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()
	c.maxTokens = config.GetMaxTokens()
	return nil
}

func (c *GoogleGenAIClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	// Available models are at https://ai.google.dev/models e.g.gemini-pro.
	model := c.client.GenerativeModel(c.model)
	model.SetTemperature(c.temperature)
	model.SetTopP(c.topP)
	model.SetMaxOutputTokens(int32(c.maxTokens))

	// Google AI SDK is capable of different inputs than just text, for now set explicit text prompt type.
	// Similarly, we could stream the response. For now k8sgpt does not support streaming.
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 {
		if resp.PromptFeedback.BlockReason == genai.BlockReasonSafety {
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

	if got.CitationMetadata != nil && len(got.CitationMetadata.CitationSources) > 0 {
		output += "Citations:\n"
		for _, source := range got.CitationMetadata.CitationSources {
			// TODO(bwplotka): Give details around what exactly words could be attributed to the citation.
			output += fmt.Sprintf("* %s, %s\n", *source.URI, source.License)
		}
	}
	return output, nil
}

func (c *GoogleGenAIClient) GetName() string {
	return googleAIClientName
}

func (c *GoogleGenAIClient) Close() {
	if err := c.client.Close(); err != nil {
		color.Red("googleai client close error: %v", err)
	}
}
