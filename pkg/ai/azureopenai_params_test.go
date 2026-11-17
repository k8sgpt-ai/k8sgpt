package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// azureRequestBody configures the Azure OpenAI backend with provider, sends one completion
// to a test server and returns the JSON body of the request.
func azureRequestBody(t *testing.T, provider *AIProvider) map[string]any {
	t.Helper()

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(raw, &body))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": "ok"}, "finish_reason": "stop"},
			},
		})
	}))
	t.Cleanup(server.Close)

	provider.Model = "test-model"
	provider.Password = "test-key"
	provider.BaseURL = server.URL

	client := &AzureAIClient{}
	require.NoError(t, client.Configure(provider))
	_, err := client.GetCompletion(context.Background(), "why is my pod crashing?")
	require.NoError(t, err)

	return body
}

func TestAzureOpenAI_UsesConfiguredMaxTokensAndTopP(t *testing.T) {
	body := azureRequestBody(t, &AIProvider{MaxTokens: 512, TopP: 0.5})

	// Sent as max_completion_tokens: reasoning deployments (o-series, gpt-5)
	// reject max_tokens.
	assert.EqualValues(t, 512, body["max_completion_tokens"])
	assert.NotContains(t, body, "max_tokens")
	assert.EqualValues(t, 0.5, body["top_p"])
}

func TestAzureOpenAI_OmitsMaxTokensAndTopPWhenUnset(t *testing.T) {
	body := azureRequestBody(t, &AIProvider{})

	assert.NotContains(t, body, "max_tokens")
	assert.NotContains(t, body, "max_completion_tokens")
	assert.NotContains(t, body, "top_p")
}

func TestAzureOpenAI_RetriesWithoutSamplingParamsWhenRejected(t *testing.T) {
	var bodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var body map[string]any
		require.NoError(t, json.Unmarshal(raw, &body))
		bodies = append(bodies, body)
		w.Header().Set("Content-Type", "application/json")
		if len(bodies) == 1 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{"message": "Unsupported value: 'top_p'", "type": "invalid_request_error", "param": "top_p"},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": "ok"}, "finish_reason": "stop"},
			},
		})
	}))
	t.Cleanup(server.Close)

	client := &AzureAIClient{}
	require.NoError(t, client.Configure(&AIProvider{Model: "test-model", Password: "test-key", BaseURL: server.URL, TopP: 0.5}))
	got, err := client.GetCompletion(context.Background(), "why is my pod crashing?")
	require.NoError(t, err)

	assert.Equal(t, "ok", got)
	require.Len(t, bodies, 2)
	assert.EqualValues(t, 0.5, bodies[0]["top_p"])
	assert.NotContains(t, bodies[1], "top_p")
}

// azureRejectingServer answers 400 for every request carrying a field in
// rejected and 200 otherwise, and records the request bodies.
func azureRejectingServer(t *testing.T, rejected string, bodies *[]map[string]any) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var body map[string]any
		require.NoError(t, json.Unmarshal(raw, &body))
		*bodies = append(*bodies, body)
		w.Header().Set("Content-Type", "application/json")
		if _, ok := body[rejected]; ok {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{"message": "Unsupported parameter: '" + rejected + "'", "type": "invalid_request_error", "param": rejected},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"role": "assistant", "content": "ok"}, "finish_reason": "stop"},
			},
		})
	}))
	t.Cleanup(server.Close)
	return server
}

func TestAzureOpenAI_RetriesWithMaxTokensWhenMaxCompletionTokensRejected(t *testing.T) {
	var bodies []map[string]any
	server := azureRejectingServer(t, "max_completion_tokens", &bodies)

	client := &AzureAIClient{}
	require.NoError(t, client.Configure(&AIProvider{Model: "test-model", Password: "test-key", BaseURL: server.URL, MaxTokens: 512}))
	got, err := client.GetCompletion(context.Background(), "why is my pod crashing?")
	require.NoError(t, err)

	assert.Equal(t, "ok", got)
	require.Len(t, bodies, 2)
	assert.EqualValues(t, 512, bodies[0]["max_completion_tokens"])
	assert.EqualValues(t, 512, bodies[1]["max_tokens"])
	assert.NotContains(t, bodies[1], "max_completion_tokens")
}

func TestAzureOpenAI_RetriesWithMaxCompletionTokensWhenMaxTokensRejected(t *testing.T) {
	var bodies []map[string]any
	server := azureRejectingServer(t, "max_tokens", &bodies)

	client := &AzureAIClient{}
	require.NoError(t, client.Configure(&AIProvider{Model: "test-model", Password: "test-key", BaseURL: server.URL, MaxTokens: 512}))
	// A server that rejects max_tokens accepts the first (max_completion_tokens) request.
	got, err := client.GetCompletion(context.Background(), "why is my pod crashing?")
	require.NoError(t, err)

	assert.Equal(t, "ok", got)
	require.Len(t, bodies, 1)
	assert.EqualValues(t, 512, bodies[0]["max_completion_tokens"])
}
