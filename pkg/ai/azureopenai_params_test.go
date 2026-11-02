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

	assert.EqualValues(t, 512, body["max_tokens"])
	assert.EqualValues(t, 0.5, body["top_p"])
}

func TestAzureOpenAI_OmitsMaxTokensAndTopPWhenUnset(t *testing.T) {
	body := azureRequestBody(t, &AIProvider{})

	assert.NotContains(t, body, "max_tokens")
	assert.NotContains(t, body, "top_p")
}
