/*
Copyright 2026 The K8sGPT Authors.
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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type customRestMockConfig struct {
	baseURL       string
	password      string
	model         string
	proxyEndpoint string
	temperature   float32
	topP          float32
	topK          int32
}

func (m *customRestMockConfig) GetPassword() string             { return m.password }
func (m *customRestMockConfig) GetModel() string                { return m.model }
func (m *customRestMockConfig) GetBaseURL() string              { return m.baseURL }
func (m *customRestMockConfig) GetProxyEndpoint() string        { return m.proxyEndpoint }
func (m *customRestMockConfig) GetEndpointName() string         { return "" }
func (m *customRestMockConfig) GetEngine() string               { return "" }
func (m *customRestMockConfig) GetTemperature() float32         { return m.temperature }
func (m *customRestMockConfig) GetProviderRegion() string       { return "" }
func (m *customRestMockConfig) GetTopP() float32                { return m.topP }
func (m *customRestMockConfig) GetTopK() int32                  { return m.topK }
func (m *customRestMockConfig) GetMaxTokens() int               { return 0 }
func (m *customRestMockConfig) GetStopSequences() []string      { return nil }
func (m *customRestMockConfig) GetProviderId() string           { return "" }
func (m *customRestMockConfig) GetCompartmentId() string        { return "" }
func (m *customRestMockConfig) GetOrganizationId() string       { return "" }
func (m *customRestMockConfig) GetAzureAPIType() string         { return "" }
func (m *customRestMockConfig) GetAzureAPIVersion() string      { return "" }
func (m *customRestMockConfig) GetCustomHeaders() []http.Header { return nil }

func TestCustomRestClient_GetName(t *testing.T) {
	c := &CustomRestClient{}
	assert.Equal(t, CustomRestClientName, c.GetName())
}

func TestCustomRestClient_Configure(t *testing.T) {
	t.Run("defaults when empty", func(t *testing.T) {
		c := &CustomRestClient{}
		err := c.Configure(&customRestMockConfig{})
		require.NoError(t, err)
		assert.Equal(t, defaultBaseURL, c.base.String())
		assert.Equal(t, defaultModel, c.model)
		assert.NotNil(t, c.client)
	})

	t.Run("explicit configuration", func(t *testing.T) {
		c := &CustomRestClient{}
		err := c.Configure(&customRestMockConfig{
			baseURL:     "http://custom-ai.internal:8080/v1/completions",
			password:    "secret-token",
			model:       "my-custom-model",
			temperature: 0.7,
			topP:        0.9,
			topK:        40,
		})
		require.NoError(t, err)
		assert.Equal(t, "http://custom-ai.internal:8080/v1/completions", c.base.String())
		assert.Equal(t, "secret-token", c.token)
		assert.Equal(t, "my-custom-model", c.model)
		assert.Equal(t, float32(0.7), c.temperature)
		assert.Equal(t, float32(0.9), c.topP)
		assert.Equal(t, int32(40), c.topK)
	})

	t.Run("with proxy endpoint", func(t *testing.T) {
		c := &CustomRestClient{}
		err := c.Configure(&customRestMockConfig{
			proxyEndpoint: "http://proxy.internal:3128",
		})
		require.NoError(t, err)
		assert.NotNil(t, c.client)
		assert.NotNil(t, c.client.Transport)
	})

	t.Run("invalid baseURL error", func(t *testing.T) {
		c := &CustomRestClient{}
		err := c.Configure(&customRestMockConfig{
			baseURL: "://invalid-url",
		})
		require.Error(t, err)
	})

	t.Run("invalid proxy endpoint error", func(t *testing.T) {
		c := &CustomRestClient{}
		err := c.Configure(&customRestMockConfig{
			proxyEndpoint: "://invalid-proxy",
		})
		require.Error(t, err)
	})
}

func TestCustomRestClient_GetCompletion_PlainText(t *testing.T) {
	var receivedRequest CustomRestRequest
	var receivedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &receivedRequest))

		resp := CustomRestResponse{
			Model:     "custom-model",
			CreatedAt: time.Now(),
			Response:  "Here is the pod analysis response",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := &CustomRestClient{}
	err := c.Configure(&customRestMockConfig{
		baseURL:  server.URL,
		password: "test-auth-token",
		model:    "custom-model",
	})
	require.NoError(t, err)

	resp, err := c.GetCompletion(context.Background(), "Why is pod test-pod CrashLoopBackOff?")
	require.NoError(t, err)
	assert.Equal(t, "Here is the pod analysis response", resp)
	assert.Equal(t, "Bearer test-auth-token", receivedAuth)
	assert.Equal(t, "custom-model", receivedRequest.Model)
	assert.Equal(t, "Why is pod test-pod CrashLoopBackOff?", receivedRequest.Prompt)
	assert.Equal(t, "Why is pod test-pod CrashLoopBackOff?", receivedRequest.Options["message"])
}

func TestCustomRestClient_GetCompletion_PlainTextWithNewlinesAndTabs(t *testing.T) {
	var receivedRequest CustomRestRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &receivedRequest))

		resp := CustomRestResponse{
			Model:    "custom-model",
			Response: "Analysis of multiline error",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := &CustomRestClient{}
	require.NoError(t, c.Configure(&customRestMockConfig{baseURL: server.URL}))

	rawPrompt := "Error log:\n\tBack-off 5m0s restarting failed container=web\nExit status 1"
	resp, err := c.GetCompletion(context.Background(), rawPrompt)
	require.NoError(t, err)
	assert.Equal(t, "Analysis of multiline error", resp)
	assert.Equal(t, rawPrompt, receivedRequest.Prompt)
	assert.Equal(t, rawPrompt, receivedRequest.Options["message"])
}

func TestCustomRestClient_GetCompletion_StructuredJSON(t *testing.T) {
	var receivedRequest CustomRestRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &receivedRequest))

		resp := CustomRestResponse{
			Model:    "custom-model",
			Response: "Structured analysis explanation",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := &CustomRestClient{}
	require.NoError(t, c.Configure(&customRestMockConfig{baseURL: server.URL}))

	structuredPrompt := `{"language":"spanish","message":"Error in Deployment","prompt":"Simplify this error in spanish"}`
	resp, err := c.GetCompletion(context.Background(), structuredPrompt)
	require.NoError(t, err)
	assert.Equal(t, "Structured analysis explanation", resp)
	assert.Equal(t, "Simplify this error in spanish", receivedRequest.Prompt)
	assert.Equal(t, "Error in Deployment", receivedRequest.Options["message"])
	assert.Equal(t, "spanish", receivedRequest.Options["language"])
}

func TestCustomRestClient_GetCompletion_JSONOnlyMessage(t *testing.T) {
	var receivedRequest CustomRestRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &receivedRequest))

		resp := CustomRestResponse{
			Response: "Done",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := &CustomRestClient{}
	require.NoError(t, c.Configure(&customRestMockConfig{baseURL: server.URL}))

	jsonOnlyMessage := `{"message":"Service port mismatch"}`
	resp, err := c.GetCompletion(context.Background(), jsonOnlyMessage)
	require.NoError(t, err)
	assert.Equal(t, "Done", resp)
	assert.Equal(t, "Service port mismatch", receivedRequest.Prompt)
	assert.Equal(t, "Service port mismatch", receivedRequest.Options["message"])
}

func TestCustomRestClient_GetCompletion_JSONOnlyPrompt(t *testing.T) {
	var receivedRequest CustomRestRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &receivedRequest))

		resp := CustomRestResponse{
			Response: "Done",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := &CustomRestClient{}
	require.NoError(t, c.Configure(&customRestMockConfig{baseURL: server.URL}))

	jsonOnlyPrompt := `{"prompt":"Explain why ingress path is 404"}`
	resp, err := c.GetCompletion(context.Background(), jsonOnlyPrompt)
	require.NoError(t, err)
	assert.Equal(t, "Done", resp)
	assert.Equal(t, "Explain why ingress path is 404", receivedRequest.Prompt)
	assert.Equal(t, "Explain why ingress path is 404", receivedRequest.Options["message"])
}

func TestCustomRestClient_GetCompletion_JSONWithUnescapedControlChars(t *testing.T) {
	var receivedRequest CustomRestRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &receivedRequest))

		resp := CustomRestResponse{
			Response: "Recovered",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := &CustomRestClient{}
	require.NoError(t, c.Configure(&customRestMockConfig{baseURL: server.URL}))

	// JSON payload that contains literal unescaped newline and tab inside string value
	jsonWithRawControlChars := "{\"language\":\"english\",\"message\":\"Error line 1\n\tline 2\",\"prompt\":\"Explain\"}"
	resp, err := c.GetCompletion(context.Background(), jsonWithRawControlChars)
	require.NoError(t, err)
	assert.Equal(t, "Recovered", resp)
	assert.Equal(t, "Explain", receivedRequest.Prompt)
	assert.Contains(t, receivedRequest.Options["message"], "Error line 1")
}

func TestCustomRestClient_GetCompletion_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "model context length exceeded", http.StatusBadRequest)
	}))
	defer server.Close()

	c := &CustomRestClient{}
	require.NoError(t, c.Configure(&customRestMockConfig{baseURL: server.URL}))

	_, err := c.GetCompletion(context.Background(), "test prompt")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Request Error, StatusCode: 400")
	assert.Contains(t, err.Error(), "model context length exceeded")
}

func TestCustomRestClient_GetCompletion_InvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not a json response"))
	}))
	defer server.Close()

	c := &CustomRestClient{}
	require.NoError(t, c.Configure(&customRestMockConfig{baseURL: server.URL}))

	_, err := c.GetCompletion(context.Background(), "test prompt")
	require.Error(t, err)
}

func TestCustomRestClient_GetCompletion_JSONWithoutMessageOrPrompt(t *testing.T) {
	tests := []struct {
		name   string
		prompt string
	}{
		{
			name:   "arbitrary json object",
			prompt: `{"foo":"bar"}`,
		},
		{
			name:   "json null",
			prompt: "null",
		},
		{
			name:   "json array",
			prompt: `["item1","item2"]`,
		},
		{
			name:   "json number",
			prompt: "12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedRequest CustomRestRequest

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.NoError(t, json.Unmarshal(body, &receivedRequest))

				resp := CustomRestResponse{
					Response: "Handled non-prompt JSON as plain text",
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			c := &CustomRestClient{}
			require.NoError(t, c.Configure(&customRestMockConfig{baseURL: server.URL}))

			resp, err := c.GetCompletion(context.Background(), tt.prompt)
			require.NoError(t, err)
			assert.Equal(t, "Handled non-prompt JSON as plain text", resp)
			assert.Equal(t, tt.prompt, receivedRequest.Prompt)
			assert.Equal(t, tt.prompt, receivedRequest.Options["message"])
		})
	}
}

