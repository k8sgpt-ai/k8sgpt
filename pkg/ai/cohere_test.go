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

func cohereRequestBody(t *testing.T, provider *AIProvider) map[string]any {
	t.Helper()

	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(raw, &body))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"text": "ok"})
	}))
	t.Cleanup(server.Close)

	provider.BaseURL = server.URL
	provider.Model = "command-r"
	provider.Password = "test-key"

	client := &CohereClient{}
	require.NoError(t, client.Configure(provider))
	_, err := client.GetCompletion(context.Background(), "why is my pod crashing?")
	require.NoError(t, err)
	return body
}

func TestCohereSendsConfiguredTopPAndTopK(t *testing.T) {
	body := cohereRequestBody(t, &AIProvider{TopP: 0.5, TopK: 40})

	assert.EqualValues(t, 0.5, body["p"])
	assert.EqualValues(t, 40, body["k"])
}

func TestCohereKeepsDefaultsWhenTopPAndTopKUnset(t *testing.T) {
	body := cohereRequestBody(t, &AIProvider{})

	assert.NotContains(t, body, "p")
	assert.EqualValues(t, 0, body["k"])
}
