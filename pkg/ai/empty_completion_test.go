package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A provider that answers 200 with blank content (e.g. a token limit hit by a
// reasoning model) must surface an error instead of an empty analysis.
func TestGetCompletion_EmptyContentIsAnError(t *testing.T) {
	const body = `{"choices": [{"message": {"role": "assistant", "content": "  "}, "finish_reason": "length"}]}`
	t.Setenv("AWS_BEARER_TOKEN_BEDROCK", "test-key")

	tests := []struct {
		name      string
		configure func(t *testing.T, url string) IAI
	}{
		{"azureopenai", func(t *testing.T, url string) IAI {
			c := &AzureAIClient{}
			require.NoError(t, c.Configure(&azureMockConfig{baseURL: url, engine: "gpt-4"}))
			return c
		}},
		{"groq", func(t *testing.T, url string) IAI {
			c := &GroqClient{}
			require.NoError(t, c.Configure(&litellmMockConfig{baseURL: url, model: "llama"}))
			return c
		}},
		{"litellm", func(t *testing.T, url string) IAI {
			c := &LiteLLMClient{}
			require.NoError(t, c.Configure(&litellmMockConfig{baseURL: url, model: "gpt-4o"}))
			return c
		}},
		{"bedrockmantle", func(t *testing.T, url string) IAI {
			c := &AmazonBedrockMantleClient{}
			require.NoError(t, c.Configure(&AIProvider{Model: "m", BaseURL: url, ProviderRegion: "us-east-1"}))
			return c
		}},
		{"gpt2giga", func(t *testing.T, url string) IAI {
			return &Gpt2GigaClient{client: http.DefaultClient, baseURL: url}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(body))
			}))
			defer server.Close()

			out, err := tt.configure(t, server.URL).GetCompletion(context.Background(), "prompt")

			require.Error(t, err)
			assert.Empty(t, out)
			assert.Contains(t, err.Error(), "empty completion")
		})
	}
}

func TestBedrockMantleGetCompletion_NoChoices(t *testing.T) {
	t.Setenv("AWS_BEARER_TOKEN_BEDROCK", "test-key")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices": []}`))
	}))
	defer server.Close()

	c := &AmazonBedrockMantleClient{}
	require.NoError(t, c.Configure(&AIProvider{Model: "m", BaseURL: server.URL, ProviderRegion: "us-east-1"}))

	out, err := c.GetCompletion(context.Background(), "prompt")

	require.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, err.Error(), "no completion choices")
}
