package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ollama, Cohere and Huggingface used to return a blank completion as if it
// were a successful analysis (and the caller cached it).
func TestGetCompletion_EmptyContentIsAnError_OllamaCohere(t *testing.T) {
	t.Run("ollama", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.URL.Path == "/api/tags" {
				_, _ = w.Write([]byte(`{"models": [{"name": "m:latest"}]}`))
				return
			}
			_, _ = w.Write([]byte(`{"model": "m", "response": "  ", "done": true, "done_reason": "length"}`))
		}))
		defer server.Close()

		c := &OllamaClient{}
		require.NoError(t, c.Configure(&mockIAIConfig{baseURL: server.URL, model: "m", temperature: 0.7, topP: 1.0}))

		out, err := c.GetCompletion(context.Background(), "prompt")

		require.Error(t, err)
		assert.Empty(t, out)
		assert.Contains(t, err.Error(), "empty completion")
	})

	t.Run("cohere", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"text": "  ", "generation_id": "g", "finish_reason": "MAX_TOKENS"}`))
		}))
		defer server.Close()

		c := &CohereClient{}
		require.NoError(t, c.Configure(&mockIAIConfig{baseURL: server.URL, model: "command", temperature: 0.7, topP: 1.0}))

		out, err := c.GetCompletion(context.Background(), "prompt")

		require.Error(t, err)
		assert.Empty(t, out)
		assert.Contains(t, err.Error(), "empty completion")
	})
}
