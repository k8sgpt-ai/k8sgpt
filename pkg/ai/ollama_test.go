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
	// Create a mock server that returns an empty model list
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"models": []}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Ensure the auto-pull env var is explicitly not set for this test
	os.Unsetenv("K8SGPT_OLLAMA_AUTO_PULL")

	// Initialize the OllamaClient using the mock server
	client := &OllamaClient{}
	config := &mockIAIConfig{
		baseURL:     server.URL,
		model:       "missing-model",
		temperature: 0.7,
		topP:        1.0,
	}

	err := client.Configure(config)
	assert.NoError(t, err)

	// Attempt to get completion, which should fail due to the missing model
	_, err = client.GetCompletion(context.Background(), "Hello")

	// Verify that it gracefully returns an error without hanging (no TTY required)
	assert.Error(t, err)
	expectedErr := fmt.Sprintf("model '%s' is required but not installed; set K8SGPT_OLLAMA_AUTO_PULL=true to download automatically", config.model)
	assert.Contains(t, err.Error(), expectedErr)
}

// mockIAIConfig is a simple mock for IAIConfig used in tests
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
