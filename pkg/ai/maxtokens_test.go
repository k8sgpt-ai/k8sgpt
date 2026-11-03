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

// maxTokensBackends lists the OpenAI-compatible backends together with the request field
// that carries their output limit.
var maxTokensBackends = []struct {
	name   string
	client func() IAI
	field  string
}{
	{"openai", func() IAI { return &OpenAIClient{} }, "max_completion_tokens"},
	{"groq", func() IAI { return &GroqClient{} }, "max_tokens"},
	{"litellm", func() IAI { return &LiteLLMClient{} }, "max_tokens"},
	{"bedrockmantle", func() IAI { return &AmazonBedrockMantleClient{} }, "max_tokens"},
}

// requestedMaxTokens configures the backend with maxTokens, sends one completion to a test
// server and returns the output limit found in the request body.
func requestedMaxTokens(t *testing.T, client IAI, field string, maxTokens int) any {
	t.Helper()
	t.Setenv("AWS_BEARER_TOKEN_BEDROCK", "test-key")

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

	require.NoError(t, client.Configure(&AIProvider{
		Model:          "test-model",
		Password:       "test-key",
		BaseURL:        server.URL,
		ProviderRegion: "us-east-1",
		MaxTokens:      maxTokens,
	}))
	_, err := client.GetCompletion(context.Background(), "why is my pod crashing?")
	require.NoError(t, err)

	return body[field]
}

func TestOpenAICompatibleBackends_UseConfiguredMaxTokens(t *testing.T) {
	for _, backend := range maxTokensBackends {
		t.Run(backend.name, func(t *testing.T) {
			got := requestedMaxTokens(t, backend.client(), backend.field, 512)

			assert.EqualValues(t, 512, got)
		})
	}
}

func TestOpenAICompatibleBackends_DefaultMaxTokensWhenUnset(t *testing.T) {
	for _, backend := range maxTokensBackends {
		t.Run(backend.name, func(t *testing.T) {
			got := requestedMaxTokens(t, backend.client(), backend.field, 0)

			assert.EqualValues(t, maxToken, got)
		})
	}
}
