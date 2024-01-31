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

func TestNewClient(t *testing.T) {
	// Test with a known provider
	client := NewClient("OpenAI")
	assert.IsType(t, &OpenAIClient{}, client)

	// Test with an unknown provider
	client = NewClient("unknown")
	assert.IsType(t, &OpenAIClient{}, client) // default client is OpenAIClient
}

func TestAIProvider(t *testing.T) {
	// Create an AIProvider
	provider := &AIProvider{
		Name:           "test",
		Model:          "model",
		Password:       "password",
		BaseURL:        "http://localhost",
		EndpointName:   "endpoint",
		Engine:         "engine",
		Temperature:    0.5,
		ProviderRegion: "region",
		TopP:           0.5,
		MaxTokens:      100,
	}

	// Test the GetBaseURL method
	assert.Equal(t, "http://localhost", provider.GetBaseURL())

	// Test the GetEndpointName method
	assert.Equal(t, "endpoint", provider.GetEndpointName())

	// Test the GetTopP method
	assert.Equal(t, float32(0.5), provider.GetTopP())

	// Test the GetMaxTokens method
	assert.Equal(t, 100, provider.GetMaxTokens())

	// Test the GetPassword method
	assert.Equal(t, "password", provider.GetPassword())

	// Test the GetModel method
	assert.Equal(t, "model", provider.GetModel())

	// Test the GetEngine method
	assert.Equal(t, "engine", provider.GetEngine())

	// Test the GetTemperature method
	assert.Equal(t, float32(0.5), provider.GetTemperature())

	// Test the GetProviderRegion method
	assert.Equal(t, "region", provider.GetProviderRegion())
}

func TestNeedPassword(t *testing.T) {
	// Test with a passwordless provider
	assert.False(t, NeedPassword("localai"))
	assert.False(t, NeedPassword("amazonsagemaker"))
	assert.False(t, NeedPassword("amazonbedrock"))

	// Test with a provider that needs a password
	assert.True(t, NeedPassword("OpenAI"))

	// Test with an unknown provider
	assert.True(t, NeedPassword("unknown"))
}
