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
	"testing"

	vertexgenai "cloud.google.com/go/vertexai/genai"
	googlegenai "github.com/google/generative-ai-go/genai"
	"github.com/stretchr/testify/assert"
)

func TestGoogleAICandidateText(t *testing.T) {
	t.Run("joins the text parts", func(t *testing.T) {
		got, err := googleAICandidateText(&googlegenai.Candidate{
			Content: &googlegenai.Content{Parts: []googlegenai.Part{googlegenai.Text("a"), googlegenai.Text("b")}},
		})
		assert.NoError(t, err)
		assert.Equal(t, "a\nb\n", got)
	})

	t.Run("a candidate without content is an error", func(t *testing.T) {
		_, err := googleAICandidateText(&googlegenai.Candidate{
			FinishReason: googlegenai.FinishReasonOther,
			SafetyRatings: []*googlegenai.SafetyRating{
				{Category: googlegenai.HarmCategoryHarassment, Probability: googlegenai.HarmProbabilityHigh, Blocked: true},
				{Category: googlegenai.HarmCategoryDangerousContent, Probability: googlegenai.HarmProbabilityLow},
			},
		})
		assert.ErrorContains(t, err, "finish reason: FinishReasonOther")
		assert.ErrorContains(t, err, "HarmCategoryHarassment")
		assert.NotContains(t, err.Error(), "HarmCategoryDangerousContent")
	})

	t.Run("a candidate with empty content is an error", func(t *testing.T) {
		_, err := googleAICandidateText(&googlegenai.Candidate{
			Content:      &googlegenai.Content{},
			FinishReason: googlegenai.FinishReasonMaxTokens,
		})
		assert.ErrorContains(t, err, "finish reason: FinishReasonMaxTokens")
	})
}

func TestVertexAICandidateText(t *testing.T) {
	t.Run("joins the text parts", func(t *testing.T) {
		got, err := vertexAICandidateText(&vertexgenai.Candidate{
			Content: &vertexgenai.Content{Parts: []vertexgenai.Part{vertexgenai.Text("a"), vertexgenai.Text("b")}},
		})
		assert.NoError(t, err)
		assert.Equal(t, "a\nb\n", got)
	})

	t.Run("a candidate without content is an error", func(t *testing.T) {
		_, err := vertexAICandidateText(&vertexgenai.Candidate{
			FinishReason:  vertexgenai.FinishReasonRecitation,
			FinishMessage: "blocked for recitation",
			SafetyRatings: []*vertexgenai.SafetyRating{
				{Category: vertexgenai.HarmCategoryHarassment, Probability: vertexgenai.HarmProbabilityHigh, Blocked: true},
			},
		})
		assert.ErrorContains(t, err, "finish reason: FinishReasonRecitation")
		assert.ErrorContains(t, err, "blocked for recitation")
		assert.ErrorContains(t, err, "HarmCategoryHarassment")
	})
}
