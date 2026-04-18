package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// azureMockConfig implements IAIConfig for Azure OpenAI tests.
type azureMockConfig struct {
	baseURL       string
	azureAPIType  string
	customHeaders []http.Header
	engine        string
	proxyEndpoint string
	organizationId string
}

func (m *azureMockConfig) GetPassword() string        { return "test-token" }
func (m *azureMockConfig) GetModel() string            { return "gpt-4" }
func (m *azureMockConfig) GetBaseURL() string          { return m.baseURL }
func (m *azureMockConfig) GetProxyEndpoint() string    { return m.proxyEndpoint }
func (m *azureMockConfig) GetEndpointName() string     { return "" }
func (m *azureMockConfig) GetEngine() string           { return m.engine }
func (m *azureMockConfig) GetTemperature() float32     { return 0.0 }
func (m *azureMockConfig) GetProviderRegion() string   { return "" }
func (m *azureMockConfig) GetTopP() float32            { return 0.0 }
func (m *azureMockConfig) GetTopK() int32              { return 0 }
func (m *azureMockConfig) GetMaxTokens() int           { return 0 }
func (m *azureMockConfig) GetStopSequences() []string  { return nil }
func (m *azureMockConfig) GetProviderId() string       { return "" }
func (m *azureMockConfig) GetCompartmentId() string    { return "" }
func (m *azureMockConfig) GetOrganizationId() string   { return m.organizationId }
func (m *azureMockConfig) GetAzureAPIType() string     { return m.azureAPIType }
func (m *azureMockConfig) GetCustomHeaders() []http.Header { return m.customHeaders }

// ---------------------------------------------------------------------------
// customHeaderRoundTripper tests
// ---------------------------------------------------------------------------

func TestCustomHeaderRoundTripper_InjectsHeaders(t *testing.T) {
	var receivedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	rt := &customHeaderRoundTripper{
		headers: []http.Header{
			{"X-Custom-One": []string{"ValueA"}},
		},
		rt: http.DefaultTransport,
	}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "ValueA", receivedHeaders.Get("X-Custom-One"))
}

func TestCustomHeaderRoundTripper_MultipleHeaders(t *testing.T) {
	var receivedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	rt := &customHeaderRoundTripper{
		headers: []http.Header{
			{"X-First": []string{"1"}},
			{"X-Second": []string{"2a", "2b"}},
		},
		rt: http.DefaultTransport,
	}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "1", receivedHeaders.Get("X-First"))
	assert.ElementsMatch(t, []string{"2a", "2b"}, receivedHeaders.Values("X-Second"))
}

func TestCustomHeaderRoundTripper_PreservesExistingHeaders(t *testing.T) {
	var receivedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	rt := &customHeaderRoundTripper{
		headers: []http.Header{
			{"X-Injected": []string{"injected-value"}},
		},
		rt: http.DefaultTransport,
	}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	req.Header.Set("X-Existing", "existing-value")

	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "existing-value", receivedHeaders.Get("X-Existing"))
	assert.Equal(t, "injected-value", receivedHeaders.Get("X-Injected"))
}

// ---------------------------------------------------------------------------
// AzureAIClient.Configure() tests
// ---------------------------------------------------------------------------

func TestAzureAIClient_Configure_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &azureMockConfig{
		baseURL: server.URL,
		engine:  "gpt-4",
	}

	client := &AzureAIClient{}
	err := client.Configure(config)
	assert.NoError(t, err)
	assert.Equal(t, "gpt-4", client.model)
}

func TestAzureAIClient_Configure_APIType(t *testing.T) {
	tests := []struct {
		name         string
		azureAPIType string
	}{
		{"empty string keeps default", ""},
		{"AZURE type", "AZURE"},
		{"AZURE_AD type", "AZURE_AD"},
		{"CLOUDFLARE_AZURE type", "CLOUDFLARE_AZURE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			config := &azureMockConfig{
				baseURL:      server.URL,
				engine:       "gpt-4",
				azureAPIType: tt.azureAPIType,
			}

			client := &AzureAIClient{}
			err := client.Configure(config)
			assert.NoError(t, err)
		})
	}
}

func TestAzureAIClient_Configure_CustomHeaders(t *testing.T) {
	headerReceived := make(chan http.Header, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerReceived <- r.Header
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"choices": [{"message": {"content": "test response"}}]}`))
	}))
	defer server.Close()

	config := &azureMockConfig{
		baseURL: server.URL,
		engine:  "gpt-4",
		customHeaders: []http.Header{
			{"X-Custom-Auth": []string{"bearer-token-123"}},
			{"X-Request-Source": []string{"k8sgpt"}},
		},
	}

	client := &AzureAIClient{}
	err := client.Configure(config)
	require.NoError(t, err)

	_, _ = client.GetCompletion(context.Background(), "test prompt")

	received := <-headerReceived
	assert.Equal(t, "bearer-token-123", received.Get("X-Custom-Auth"))
	assert.Equal(t, "k8sgpt", received.Get("X-Request-Source"))
}

func TestAzureAIClient_Configure_NoCustomHeaders_NoProxy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &azureMockConfig{
		baseURL: server.URL,
		engine:  "gpt-4",
	}

	client := &AzureAIClient{}
	err := client.Configure(config)
	assert.NoError(t, err)
}

func TestAzureAIClient_Configure_WithOrganizationId(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &azureMockConfig{
		baseURL:        server.URL,
		engine:         "gpt-4",
		organizationId: "org-12345",
	}

	client := &AzureAIClient{}
	err := client.Configure(config)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// AzureAIClient.GetCompletion() end-to-end tests
// ---------------------------------------------------------------------------

func TestAzureAIClient_GetCompletion_WithCustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify custom headers are present
		assert.Equal(t, "Value1", r.Header.Get("X-Custom-Header-1"))
		assert.ElementsMatch(t, []string{"Value2", "Value3"}, r.Header.Values("X-Custom-Header-2"))

		w.WriteHeader(http.StatusOK)
		mockResponse := `{"choices": [{"message": {"content": "azure test response"}}]}`
		_, _ = w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	config := &azureMockConfig{
		baseURL: server.URL,
		engine:  "gpt-4",
		customHeaders: []http.Header{
			{"X-Custom-Header-1": []string{"Value1"}},
			{"X-Custom-Header-2": []string{"Value2", "Value3"}},
		},
	}

	client := &AzureAIClient{}
	err := client.Configure(config)
	require.NoError(t, err)

	result, err := client.GetCompletion(context.Background(), "test prompt")
	require.NoError(t, err)
	assert.Equal(t, "azure test response", result)
}

func TestAzureAIClient_GetCompletion_WithoutCustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		mockResponse := `{"choices": [{"message": {"content": "plain response"}}]}`
		_, _ = w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	config := &azureMockConfig{
		baseURL: server.URL,
		engine:  "gpt-4",
	}

	client := &AzureAIClient{}
	err := client.Configure(config)
	require.NoError(t, err)

	result, err := client.GetCompletion(context.Background(), "test prompt")
	require.NoError(t, err)
	assert.Equal(t, "plain response", result)
}

func TestAzureAIClient_GetName(t *testing.T) {
	client := &AzureAIClient{}
	assert.Equal(t, "azureopenai", client.GetName())
}
