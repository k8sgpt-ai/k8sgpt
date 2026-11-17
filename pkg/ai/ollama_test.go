package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

	t.Setenv("K8SGPT_OLLAMA_AUTO_PULL", "")

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

	t.Setenv("K8SGPT_OLLAMA_AUTO_PULL", "true")

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

func ollamaGenerateOptions(t *testing.T, cfg *mockIAIConfig) map[string]interface{} {
	t.Helper()
	var body struct {
		Options map[string]interface{} `json:"options"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			_, _ = w.Write([]byte(`{"models": [{"name": "m:latest"}]}`))
		case "/api/generate":
			_ = json.NewDecoder(r.Body).Decode(&body)
			_, _ = w.Write([]byte(`{"response": "ok", "done": true}` + "\n"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cfg.baseURL = server.URL
	cfg.model = "m"
	client := &OllamaClient{}
	assert.NoError(t, client.Configure(cfg))
	_, err := client.GetCompletion(context.Background(), "Hello")
	assert.NoError(t, err)
	return body.Options
}

func TestOllamaSendsMaxTokensAndTopK(t *testing.T) {
	opts := ollamaGenerateOptions(t, &mockIAIConfig{temperature: 0.7, topP: 1.0, topK: 40, maxTokens: 512})

	assert.EqualValues(t, 512, opts["num_predict"])
	assert.EqualValues(t, 40, opts["top_k"])
}

func TestOllamaOmitsUnsetMaxTokensAndTopK(t *testing.T) {
	opts := ollamaGenerateOptions(t, &mockIAIConfig{temperature: 0.7, topP: 1.0})

	assert.NotContains(t, opts, "num_predict")
	assert.NotContains(t, opts, "top_k")
}

type mockIAIConfig struct {
	baseURL     string
	model       string
	temperature float32
	topP        float32
	topK        int32
	maxTokens   int
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
func (m *mockIAIConfig) GetTopK() int32 { return m.topK }
func (m *mockIAIConfig) GetMaxTokens() int { return m.maxTokens }
func (m *mockIAIConfig) GetStopSequences() []string { return nil }
func (m *mockIAIConfig) GetProviderId() string { return "" }
func (m *mockIAIConfig) GetCompartmentId() string { return "" }
func (m *mockIAIConfig) GetOrganizationId() string { return "" }
func (m *mockIAIConfig) GetAzureAPIType() string { return "" }
func (m *mockIAIConfig) GetAzureAPIVersion() string { return "" }
func (m *mockIAIConfig) GetCustomHeaders() []http.Header { return nil }
func (m *mockIAIConfig) GetProxyEndpoint() string { return "" }
