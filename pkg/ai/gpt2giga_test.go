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

func TestGpt2GigaClient_GetCompletion(t *testing.T) {
	type completionRequest struct {
		Model       string  `json:"model"`
		Temperature float32 `json:"temperature"`
		TopP        float32 `json:"top_p"`
		Messages    []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}
	type receivedRequest struct {
		method      string
		path        string
		contentType string
		body        completionRequest
		err         error
	}

	requests := make(chan receivedRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body completionRequest
		err := json.NewDecoder(r.Body).Decode(&body)
		requests <- receivedRequest{
			method:      r.Method,
			path:        r.URL.Path,
			contentType: r.Header.Get("Content-Type"),
			body:        body,
			err:         err,
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"diagnosis"}}]}`))
	}))
	defer server.Close()

	client := &Gpt2GigaClient{
		client:      server.Client(),
		baseURL:     server.URL,
		model:       "gpt2giga-test",
		temperature: 0.2,
		topP:        0.7,
	}

	completion, err := client.GetCompletion(context.Background(), "why is my pod failing?")
	require.NoError(t, err)
	assert.Equal(t, "diagnosis", completion)

	request := <-requests
	assert.NoError(t, request.err)
	assert.Equal(t, http.MethodPost, request.method)
	assert.Equal(t, "/chat/completions", request.path)
	assert.Equal(t, "application/json", request.contentType)
	assert.Equal(t, "gpt2giga-test", request.body.Model)
	assert.Equal(t, float32(0.2), request.body.Temperature)
	assert.Equal(t, float32(0.7), request.body.TopP)
	require.Len(t, request.body.Messages, 1)
	assert.Equal(t, "user", request.body.Messages[0].Role)
	assert.Equal(t, "why is my pod failing?", request.body.Messages[0].Content)
}

func TestGpt2GigaClient_GetCompletionErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantError  string
	}{
		{
			name:       "unexpected status",
			statusCode: http.StatusBadGateway,
			body:       `{}`,
			wantError:  "unexpected status code: 502",
		},
		{
			name:       "invalid response JSON",
			statusCode: http.StatusOK,
			body:       `{`,
			wantError:  "error decoding response",
		},
		{
			name:       "no choices",
			statusCode: http.StatusOK,
			body:       `{"choices":[]}`,
			wantError:  "no choices in response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := &Gpt2GigaClient{client: server.Client(), baseURL: server.URL}
			completion, err := client.GetCompletion(context.Background(), "prompt")

			assert.Empty(t, completion)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantError)
		})
	}
}
