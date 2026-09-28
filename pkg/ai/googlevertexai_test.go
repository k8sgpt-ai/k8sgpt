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

	"github.com/stretchr/testify/assert"
)

func TestGetVertexAIModelOrDefault(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  string
	}{
		{"latest stable pro is preserved", ModelGeminiProV2_5, "gemini-2.5-pro"},
		{"legacy stable pro is preserved", "gemini-1.5-pro-002", "gemini-1.5-pro-002"},
		{"legacy stable flash is preserved", "gemini-1.5-flash-002", "gemini-1.5-flash-002"},
		{"unknown model falls back to default", "does-not-exist", VERTEXAI_MODELS[0]},
		{"empty model falls back to default", "", VERTEXAI_MODELS[0]},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetVertexAIModelOrDefault(tt.model))
		})
	}
}

func TestGetVertexAIRegionOrDefault(t *testing.T) {
	tests := []struct {
		name   string
		region string
		want   string
	}{
		{"supported region is preserved", US_West_4, "us-west4"},
		{"unknown region falls back to default", "moon-base-1", VERTEXAI_DEFAULT_REGION},
		{"empty region falls back to default", "", VERTEXAI_DEFAULT_REGION},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetVertexAIRegionOrDefault(tt.region))
		})
	}
}
