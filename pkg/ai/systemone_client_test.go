package ai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockIAIConfig struct {
	model   string
	baseURL string
}

func (m *mockIAIConfig) GetPassword() string             { return "" }
func (m *mockIAIConfig) GetModel() string                { return m.model }
func (m *mockIAIConfig) GetBaseURL() string              { return m.baseURL }
func (m *mockIAIConfig) GetProxyEndpoint() string        { return "" }
func (m *mockIAIConfig) GetEndpointName() string         { return "" }
func (m *mockIAIConfig) GetEngine() string               { return "" }
func (m *mockIAIConfig) GetTemperature() float32         { return 0 }
func (m *mockIAIConfig) GetProviderRegion() string       { return "" }
func (m *mockIAIConfig) GetTopP() float32                { return 0 }
func (m *mockIAIConfig) GetTopK() int32                  { return 0 }
func (m *mockIAIConfig) GetMaxTokens() int               { return 0 }
func (m *mockIAIConfig) GetStopSequences() []string      { return nil }
func (m *mockIAIConfig) GetProviderId() string           { return "" }
func (m *mockIAIConfig) GetCompartmentId() string        { return "" }
func (m *mockIAIConfig) GetOrganizationId() string       { return "" }
func (m *mockIAIConfig) GetAzureAPIType() string         { return "" }
func (m *mockIAIConfig) GetAzureAPIVersion() string      { return "" }
func (m *mockIAIConfig) GetCustomHeaders() []http.Header { return nil }

func TestSystemOneClient_ClassifyAction(t *testing.T) {
	// 1. Create a mock HTTP server to act like the Python Laya server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/classify", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		response := ai.ActionClassification{
			Decision:        "RestartPod",
			Confidence:      0.99,
			SuggestedAction: "Mock action",
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	// 2. Configure the SystemOne client
	client := &ai.SystemOneClient{}
	config := &mockIAIConfig{
		model:   "laya",
		baseURL: ts.URL, // Point to our mock server
	}
	err := client.Configure(config)
	require.NoError(t, err)

	// 3. Run the ClassifyAction method
	ctx := context.Background()
	categories := []string{"RestartPod", "Ignore"}

	result, err := client.ClassifyAction(ctx, "mock prompt", categories)

	// 4. Verify results
	require.NoError(t, err)
	assert.Equal(t, "RestartPod", result.Decision)
	assert.InDelta(t, 0.99, result.Confidence, 0.001)
	assert.Equal(t, "Mock action", result.SuggestedAction)
}

func TestSystemOneClient_GetCompletion_Error(t *testing.T) {
	client := &ai.SystemOneClient{}
	_, err := client.GetCompletion(context.Background(), "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deterministic classifiers")
}

func TestSystemOneClient_GetName(t *testing.T) {
	client := &ai.SystemOneClient{}
	assert.Equal(t, "systemone", client.GetName())
}

func TestSystemOneClient_ClassifyAction_EmptyModel(t *testing.T) {
	client := &ai.SystemOneClient{}
	err := client.Configure(&mockIAIConfig{model: "", baseURL: "http://localhost"})
	require.NoError(t, err)

	_, err = client.ClassifyAction(context.Background(), "prompt", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "specific model")
}

func TestSystemOneClient_ClassifyAction_EmptyBaseURL(t *testing.T) {
	client := &ai.SystemOneClient{}
	err := client.Configure(&mockIAIConfig{model: "laya", baseURL: ""})
	require.NoError(t, err)

	_, err = client.ClassifyAction(context.Background(), "prompt", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "base URL must be configured")
}

func TestSystemOneClient_ClassifyAction_BadURL(t *testing.T) {
	client := &ai.SystemOneClient{}
	// Control character in URL causes http.NewRequest to fail
	err := client.Configure(&mockIAIConfig{model: "laya", baseURL: "http://\x7flocalhost"})
	require.NoError(t, err)

	_, err = client.ClassifyAction(context.Background(), "prompt", []string{})
	assert.Error(t, err)
}

func TestSystemOneClient_ClassifyAction_HttpError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := &ai.SystemOneClient{}
	err := client.Configure(&mockIAIConfig{model: "laya", baseURL: ts.URL})
	require.NoError(t, err)

	_, err = client.ClassifyAction(context.Background(), "prompt", []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get classification")
}

func TestSystemOneClient_ClassifyAction_BadJSONResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{bad json`))
	}))
	defer ts.Close()

	client := &ai.SystemOneClient{}
	err := client.Configure(&mockIAIConfig{model: "laya", baseURL: ts.URL})
	require.NoError(t, err)

	_, err = client.ClassifyAction(context.Background(), "prompt", []string{})
	assert.Error(t, err)
}

func TestSystemOneClient_ClassifyAction_ClientDoError(t *testing.T) {
	client := &ai.SystemOneClient{}
	err := client.Configure(&mockIAIConfig{model: "laya", baseURL: "http://invalid-host-that-does-not-exist.local"})
	require.NoError(t, err)

	_, err = client.ClassifyAction(context.Background(), "prompt", []string{})
	assert.Error(t, err)
}
