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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// litellmMockConfig implements IAIConfig for LiteLLM backend tests.
type litellmMockConfig struct {
	baseURL       string
	password      string
	model         string
	customHeaders []http.Header
}

func (m *litellmMockConfig) GetPassword() string             { return m.password }
func (m *litellmMockConfig) GetModel() string                { return m.model }
func (m *litellmMockConfig) GetBaseURL() string              { return m.baseURL }
func (m *litellmMockConfig) GetProxyEndpoint() string        { return "" }
func (m *litellmMockConfig) GetEndpointName() string         { return "" }
func (m *litellmMockConfig) GetEngine() string               { return "" }
func (m *litellmMockConfig) GetTemperature() float32         { return 0.0 }
func (m *litellmMockConfig) GetProviderRegion() string       { return "" }
func (m *litellmMockConfig) GetTopP() float32                { return 0.0 }
func (m *litellmMockConfig) GetTopK() int32                  { return 0 }
func (m *litellmMockConfig) GetMaxTokens() int               { return 0 }
func (m *litellmMockConfig) GetStopSequences() []string      { return nil }
func (m *litellmMockConfig) GetProviderId() string           { return "" }
func (m *litellmMockConfig) GetCompartmentId() string        { return "" }
func (m *litellmMockConfig) GetOrganizationId() string       { return "" }
func (m *litellmMockConfig) GetAzureAPIType() string         { return "" }
func (m *litellmMockConfig) GetAzureAPIVersion() string      { return "" }
func (m *litellmMockConfig) GetCustomHeaders() []http.Header { return m.customHeaders }

func TestLiteLLMClient_GetName(t *testing.T) {
	c := &LiteLLMClient{}
	assert.Equal(t, "litellm", c.GetName())
}

func TestLiteLLMClient_ConfigureDefaultsToProxyBaseURL(t *testing.T) {
	c := &LiteLLMClient{}
	err := c.Configure(&litellmMockConfig{model: "gpt-4o"})
	require.NoError(t, err)
	require.NotNil(t, c.client)
	assert.Equal(t, "gpt-4o", c.model)
}

func TestLiteLLMClient_IsRegisteredAndPasswordless(t *testing.T) {
	// Resolvable by name through the backend registry.
	assert.IsType(t, &LiteLLMClient{}, NewClient(liteLLMClientName))
	assert.Contains(t, Backends, liteLLMClientName)
	// Self-hosted proxies are commonly keyless, so no password is required.
	assert.False(t, NeedPassword(liteLLMClientName))
}

func TestLiteLLMClient_GetCompletion(t *testing.T) {
	var gotAuth string
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": "the cluster is healthy"}},
			},
		})
	}))
	defer server.Close()

	c := &LiteLLMClient{}
	err := c.Configure(&litellmMockConfig{
		baseURL:  server.URL + "/v1",
		password: "sk-litellm-virtual-key",
		model:    "gpt-4o",
	})
	require.NoError(t, err)

	resp, err := c.GetCompletion(context.Background(), "why is my pod crashing?")
	require.NoError(t, err)
	assert.Equal(t, "the cluster is healthy", resp)
	assert.Equal(t, "Bearer sk-litellm-virtual-key", gotAuth)
	assert.Equal(t, "/v1/chat/completions", gotPath)
}
