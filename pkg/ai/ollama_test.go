package ai

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOllamaGetCompletionMissingModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"models": []}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	os.Unsetenv("K8SGPT_OLLAMA_AUTO_PULL")

	client := &OllamaClient{}
	config := &mockIAIConfig{
		baseURL:     server.URL,
		model:       "missing-model",
		temperature: 0.7,
		topP:        1.0,
	}

	err := client.Configure(config)
	assert.NoError(t, err)

	_, err = client.GetCompletion(context.Background(), "Hello")

	assert.Error(t, err)
	expectedErr := fmt.Sprintf("model '%s' is required but not installed; set K8SGPT_OLLAMA_AUTO_PULL=true to download automatically", config.model)
	assert.Contains(t, err.Error(), expectedErr)
}

func TestOllamaGetCompletionAutoPull(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"models": []}`))
			return
		}
		if r.URL.Path == "/api/pull" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status": "success", "total": 100, "completed": 50}` + "\n" + `{"status": "success", "total": 100, "completed": 100}` + "\n"))
			return
		}
		if r.URL.Path == "/api/generate" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"response": "Here is a response", "done": true}` + "\n"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	os.Setenv("K8SGPT_OLLAMA_AUTO_PULL", "true")
	defer os.Unsetenv("K8SGPT_OLLAMA_AUTO_PULL")

	client := &OllamaClient{}
	config := &mockIAIConfig{
		baseURL:     server.URL,
		model:       "missing-model",
		temperature: 0.7,
		topP:        1.0,
	}

	err := client.Configure(config)
	assert.NoError(t, err)

	resp, err := client.GetCompletion(context.Background(), "Hello")

	assert.NoError(t, err)
	assert.Equal(t, "Here is a response", resp)
}

func TestOllamaGetCompletionModelExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"models": [{"name": "existing-model"}]}`))
			return
		}
		if r.URL.Path == "/api/generate" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"response": "Here is a response for existing model", "done": true}` + "\n"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &OllamaClient{}
	config := &mockIAIConfig{
		baseURL:     server.URL,
		model:       "existing-model",
		temperature: 0.7,
		topP:        1.0,
	}

	err := client.Configure(config)
	assert.NoError(t, err)

	resp, err := client.GetCompletion(context.Background(), "Hello")

	assert.NoError(t, err)
	assert.Equal(t, "Here is a response for existing model", resp)
}

type mockIAIConfig struct {
	baseURL     string
	model       string
	temperature float32
	topP        float32
}

func (m *mockIAIConfig) GetBaseURL() string {
	return m.baseURL
}

func (m *mockIAIConfig) GetModel() string {
	return m.model
}

func (m *mockIAIConfig) GetTemperature() float32 {
	return m.temperature
}

func (m *mockIAIConfig) GetTopP() float32 {
	return m.topP
}

func (m *mockIAIConfig) GetPassword() string { return "" }
func (m *mockIAIConfig) GetEndpointName() string { return "" }
func (m *mockIAIConfig) GetEngine() string { return "" }
func (m *mockIAIConfig) GetProviderRegion() string { return "" }
func (m *mockIAIConfig) GetTopK() int32 { return 0 }
func (m *mockIAIConfig) GetMaxTokens() int { return 0 }
func (m *mockIAIConfig) GetStopSequences() []string { return nil }
func (m *mockIAIConfig) GetProviderId() string { return "" }
func (m *mockIAIConfig) GetCompartmentId() string { return "" }
func (m *mockIAIConfig) GetOrganizationId() string { return "" }
func (m *mockIAIConfig) GetAzureAPIType() string { return "" }
func (m *mockIAIConfig) GetAzureAPIVersion() string { return "" }
func (m *mockIAIConfig) GetCustomHeaders() []http.Header { return nil }
func (m *mockIAIConfig) GetProxyEndpoint() string { return "" }
