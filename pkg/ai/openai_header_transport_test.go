package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock configuration
type mockConfig struct {
	baseURL string
}

func (m *mockConfig) GetPassword() string {
	return ""
}

func (m *mockConfig) GetOrganizationId() string {
	return ""
}

func (m *mockConfig) GetProxyEndpoint() string {
	return ""
}

func (m *mockConfig) GetBaseURL() string {
	return m.baseURL
}

func (m *mockConfig) GetCustomHeaders() []http.Header {
	return []http.Header{
		{"X-Custom-Header-1": []string{"Value1"}},
		{"X-Custom-Header-2": []string{"Value2"}},
		{"X-Custom-Header-2": []string{"Value3"}}, // Testing multiple values for the same header
	}
}

func (m *mockConfig) GetModel() string {
	return ""
}

func (m *mockConfig) GetTemperature() float32 {
	return 0.0
}

func (m *mockConfig) GetTopP() float32 {
	return 0.0
}
func (m *mockConfig) GetCompartmentId() string {
	return ""
}

func (m *mockConfig) GetTopK() int32 {
	return 0.0
}

func (m *mockConfig) GetMaxTokens() int {
	return 0
}

func (m *mockConfig) GetEndpointName() string {
	return ""
}
func (m *mockConfig) GetEngine() string {
	return ""
}

func (m *mockConfig) GetProviderId() string {
	return ""
}

func (m *mockConfig) GetProviderRegion() string {
	return ""
}

func TestOpenAIClient_CustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Value1", r.Header.Get("X-Custom-Header-1"))
		assert.ElementsMatch(t, []string{"Value2", "Value3"}, r.Header["X-Custom-Header-2"])
		w.WriteHeader(http.StatusOK)
		// Mock response for openai completion
		mockResponse := `{"choices": [{"message": {"content": "test"}}]}`
		n, err := w.Write([]byte(mockResponse))
		if err != nil {
			t.Fatalf("error writing response: %v", err)
		}
		if n != len(mockResponse) {
			t.Fatalf("expected to write %d bytes but wrote %d bytes", len(mockResponse), n)
		}
	}))
	defer server.Close()

	config := &mockConfig{baseURL: server.URL}

	client := &OpenAIClient{}
	err := client.Configure(config)
	assert.NoError(t, err)

	// Make a completion request to trigger the headers
	ctx := context.Background()
	_, err = client.GetCompletion(ctx, "foo prompt")
	assert.NoError(t, err)
}
