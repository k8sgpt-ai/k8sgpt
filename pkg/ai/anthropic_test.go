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

type anthropicMockConfig struct {
	baseURL       string
	password      string
	model         string
	temperature   float32
	topP          float32
	topK          int32
	maxTokens     int
	stopSequences []string
	customHeaders []http.Header
}

func (m *anthropicMockConfig) GetPassword() string             { return m.password }
func (m *anthropicMockConfig) GetModel() string                { return m.model }
func (m *anthropicMockConfig) GetBaseURL() string              { return m.baseURL }
func (m *anthropicMockConfig) GetProxyEndpoint() string        { return "" }
func (m *anthropicMockConfig) GetEndpointName() string         { return "" }
func (m *anthropicMockConfig) GetEngine() string               { return "" }
func (m *anthropicMockConfig) GetTemperature() float32         { return m.temperature }
func (m *anthropicMockConfig) GetProviderRegion() string       { return "" }
func (m *anthropicMockConfig) GetTopP() float32                { return m.topP }
func (m *anthropicMockConfig) GetTopK() int32                  { return m.topK }
func (m *anthropicMockConfig) GetMaxTokens() int               { return m.maxTokens }
func (m *anthropicMockConfig) GetStopSequences() []string      { return m.stopSequences }
func (m *anthropicMockConfig) GetProviderId() string           { return "" }
func (m *anthropicMockConfig) GetCompartmentId() string        { return "" }
func (m *anthropicMockConfig) GetOrganizationId() string       { return "" }
func (m *anthropicMockConfig) GetAzureAPIType() string         { return "" }
func (m *anthropicMockConfig) GetAzureAPIVersion() string      { return "" }
func (m *anthropicMockConfig) GetCustomHeaders() []http.Header { return m.customHeaders }

func TestAnthropicClientGetCompletion(t *testing.T) {
	type requestBody struct {
		Model         string   `json:"model"`
		MaxTokens     int      `json:"max_tokens"`
		Temperature   float32  `json:"temperature"`
		TopP          float32  `json:"top_p"`
		TopK          int32    `json:"top_k"`
		StopSequences []string `json:"stop_sequences"`
		Messages      []struct {
			Role    string `json:"role"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"messages"`
	}

	// When temperature is set it takes priority; top_p must not be sent.
	t.Run("temperature takes priority over top_p", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/v1/messages", r.URL.Path)
			assert.Equal(t, "test-token", r.Header.Get("X-Api-Key"))
			assert.NotEmpty(t, r.Header.Get("Anthropic-Version"))
			assert.Equal(t, "test-value", r.Header.Get("X-Test-Header"))

			var body requestBody
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "claude-test", body.Model)
			assert.Equal(t, 1024, body.MaxTokens)
			assert.Equal(t, float32(0.1), body.Temperature)
			assert.Equal(t, float32(0), body.TopP)
			assert.Equal(t, int32(25), body.TopK)
			assert.Equal(t, []string{"STOP"}, body.StopSequences)
			require.Len(t, body.Messages, 1)
			assert.Equal(t, "user", body.Messages[0].Role)
			require.Len(t, body.Messages[0].Content, 1)
			assert.Equal(t, "text", body.Messages[0].Content[0].Type)
			assert.Equal(t, "hello cluster", body.Messages[0].Content[0].Text)

			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{"content":[{"type":"text","text":"diagnosis"}]}`))
			require.NoError(t, err)
		}))
		defer server.Close()

		client := &AnthropicClient{}
		err := client.Configure(&anthropicMockConfig{
			baseURL:       server.URL,
			password:      "test-token",
			model:         "claude-test",
			temperature:   0.1,
			topP:          0.8,
			topK:          25,
			maxTokens:     1024,
			stopSequences: []string{"STOP"},
			customHeaders: []http.Header{{"X-Test-Header": []string{"test-value"}}},
		})
		require.NoError(t, err)

		completion, err := client.GetCompletion(context.Background(), "hello cluster")
		require.NoError(t, err)
		assert.Equal(t, "diagnosis", completion)
	})

	// When temperature is zero, top_p is used instead.
	t.Run("top_p used when temperature is zero", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var body requestBody
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, float32(0), body.Temperature)
			assert.Equal(t, float32(0.9), body.TopP)

			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{"content":[{"type":"text","text":"ok"}]}`))
			require.NoError(t, err)
		}))
		defer server.Close()

		client := &AnthropicClient{}
		err := client.Configure(&anthropicMockConfig{
			baseURL:   server.URL,
			password:  "test-token",
			model:     "claude-test",
			topP:      0.9,
			maxTokens: 512,
		})
		require.NoError(t, err)

		completion, err := client.GetCompletion(context.Background(), "hello cluster")
		require.NoError(t, err)
		assert.Equal(t, "ok", completion)
	})
}

func TestAnthropicClientHonorsExplicitMessagesURLAndDefaultModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/custom/v1/messages", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"content":[{"type":"text","text":"ok"}]}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	client := &AnthropicClient{}
	err := client.Configure(&anthropicMockConfig{
		baseURL:   server.URL + "/custom/v1/messages",
		password:  "test-token",
		maxTokens: 0,
	})
	require.NoError(t, err)
	assert.Equal(t, anthropicDefaultModel, client.model)
	assert.Equal(t, anthropicDefaultMaxTokens, client.maxTokens)

	completion, err := client.GetCompletion(context.Background(), "hello")
	require.NoError(t, err)
	assert.Equal(t, "ok", completion)
}

func TestAnthropicClientReturnsStructuredErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{"error":{"type":"invalid_request_error","message":"bad prompt"}}`))
		require.NoError(t, err)
	}))
	defer server.Close()

	client := &AnthropicClient{}
	err := client.Configure(&anthropicMockConfig{baseURL: server.URL, password: "test-token"})
	require.NoError(t, err)

	_, err = client.GetCompletion(context.Background(), "hello")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad prompt")
}
