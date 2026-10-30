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

// newSamplingTestClient answers each request with the next entry of statuses/bodies
// and records the decoded request bodies.
func newSamplingTestClient(t *testing.T, statuses []int, bodies []string) (*OpenAIClient, *[]map[string]interface{}) {
	t.Helper()
	var requests []map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var got map[string]interface{}
		require.NoError(t, json.Unmarshal(raw, &got))
		i := len(requests)
		requests = append(requests, got)
		require.Less(t, i, len(statuses), "unexpected extra request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statuses[i])
		_, err = w.Write([]byte(bodies[i]))
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	client := &OpenAIClient{}
	require.NoError(t, client.Configure(&mockConfig{baseURL: server.URL}))
	client.model = "some-model"
	client.temperature = 0.7
	client.topP = 0.5
	return client, &requests
}

const okBody = `{"choices": [{"message": {"role": "assistant", "content": "ok"}}]}`

func unsupportedBody(param string) string {
	return `{"error": {"message": "Unsupported value: '` + param + `' does not support 0.7 with this model.", "type": "invalid_request_error", "param": "` + param + `", "code": "unsupported_value"}}`
}

func TestOpenAIClient_GetCompletion_RetriesWithoutUnsupportedSamplingParams(t *testing.T) {
	for _, param := range []string{"temperature", "top_p"} {
		t.Run(param, func(t *testing.T) {
			client, requests := newSamplingTestClient(t, []int{400, 200}, []string{unsupportedBody(param), okBody})

			out, err := client.GetCompletion(context.Background(), "foo prompt")

			require.NoError(t, err)
			assert.Equal(t, "ok", out)
			require.Len(t, *requests, 2)
			assert.Contains(t, (*requests)[0], "temperature")
			assert.Contains(t, (*requests)[0], "top_p")
			assert.NotContains(t, (*requests)[1], "temperature")
			assert.NotContains(t, (*requests)[1], "top_p")
		})
	}
}

func TestOpenAIClient_GetCompletion_DoesNotRetryOtherBadRequests(t *testing.T) {
	body := `{"error": {"message": "bad messages", "type": "invalid_request_error", "param": "messages", "code": "invalid_value"}}`
	client, requests := newSamplingTestClient(t, []int{400}, []string{body})

	_, err := client.GetCompletion(context.Background(), "foo prompt")

	assert.Error(t, err)
	assert.Len(t, *requests, 1)
}

func TestOpenAIClient_GetCompletion_RetriesOnlyOnce(t *testing.T) {
	client, requests := newSamplingTestClient(t, []int{400, 400}, []string{unsupportedBody("temperature"), unsupportedBody("top_p")})

	_, err := client.GetCompletion(context.Background(), "foo prompt")

	assert.Error(t, err)
	assert.Len(t, *requests, 2)
}
