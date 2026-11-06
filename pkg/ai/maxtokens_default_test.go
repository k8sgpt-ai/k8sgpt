package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// With no maxtokens in the config (zero value), these backends must fall back to maxToken
// like the OpenAI-compatible ones, instead of sending a limit of 0.

func TestAmazonBedrockConverse_DefaultMaxTokensWhenUnset(t *testing.T) {
	mock := &mockConverseClient{
		converseFunc: func(ctx context.Context, input *bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error) {
			return &bedrockruntime.ConverseOutput{
				Output: &types.ConverseOutputMemberMessage{
					Value: types.Message{
						Content: []types.ContentBlock{&types.ContentBlockMemberText{Value: "ok"}},
					},
				},
			}, nil
		},
	}
	client := &AmazonBedrockConverseClient{client: mock}
	require.NoError(t, client.Configure(&AIProvider{Model: "test-model", ProviderRegion: "us-west-2"}))

	_, err := client.GetCompletion(context.Background(), "why is my pod crashing?")
	require.NoError(t, err)

	require.NotNil(t, mock.lastInput.InferenceConfig.MaxTokens)
	assert.EqualValues(t, maxToken, *mock.lastInput.InferenceConfig.MaxTokens)
}

func TestCohere_DefaultMaxTokensWhenUnset(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(raw, &body))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"text": "ok"})
	}))
	t.Cleanup(server.Close)

	client := &CohereClient{}
	require.NoError(t, client.Configure(&AIProvider{Model: "command-r", Password: "test-key", BaseURL: server.URL}))

	_, err := client.GetCompletion(context.Background(), "why is my pod crashing?")
	require.NoError(t, err)

	assert.EqualValues(t, maxToken, body["max_tokens"])
}

func TestBackendsWithoutHTTPTestHook_DefaultMaxTokensWhenUnset(t *testing.T) {
	t.Run("amazonsagemaker", func(t *testing.T) {
		client := &SageMakerAIClient{}
		require.NoError(t, client.Configure(&AIProvider{Model: "m", ProviderRegion: "us-east-1"}))
		assert.EqualValues(t, maxToken, client.maxTokens)
	})
	t.Run("googlegenai", func(t *testing.T) {
		client := &GoogleGenAIClient{}
		require.NoError(t, client.Configure(&AIProvider{Model: "gemini-pro", Password: "test-key"}))
		assert.EqualValues(t, maxToken, client.maxTokens)
	})
}
