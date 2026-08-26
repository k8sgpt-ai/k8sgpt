package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newMockOpenAIClient(t *testing.T, body string) *OpenAIClient {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatalf("error writing response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := &OpenAIClient{}
	if err := client.Configure(&mockConfig{baseURL: server.URL}); err != nil {
		t.Fatalf("configure failed: %v", err)
	}
	return client
}

func TestOpenAIClient_GetCompletion_EmptyChoices(t *testing.T) {
	client := newMockOpenAIClient(t, `{"choices": []}`)

	out, err := client.GetCompletion(context.Background(), "foo prompt")

	assert.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, err.Error(), "empty response")
}

func TestOpenAIClient_GetCompletion_EmptyContent(t *testing.T) {
	client := newMockOpenAIClient(t, `{"choices": [{"message": {"role": "assistant", "content": "   "}, "finish_reason": "length"}]}`)

	out, err := client.GetCompletion(context.Background(), "foo prompt")

	assert.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, err.Error(), "empty completion")
	assert.Contains(t, err.Error(), "length")
}

func TestOpenAIClient_GetCompletion_NonEmptyContent(t *testing.T) {
	client := newMockOpenAIClient(t, `{"choices": [{"message": {"role": "assistant", "content": "ok"}}]}`)

	out, err := client.GetCompletion(context.Background(), "foo prompt")

	assert.NoError(t, err)
	assert.Equal(t, "ok", out)
}
