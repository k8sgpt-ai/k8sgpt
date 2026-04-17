package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBedrockMantleGetName(t *testing.T) {
	client := &AmazonBedrockMantleClient{}
	assert.Equal(t, "bedrockmantle", client.GetName())
}

func TestBedrockMantleConfigure(t *testing.T) {
	tests := []struct {
		name        string
		envRegion   string
		envToken    string
		cfg         *AIProvider
		expectError bool
		errContains string
	}{
		{
			name:      "missing region",
			envRegion: "",
			envToken:  "test-key",
			cfg: &AIProvider{
				Model: "test-model",
			},
			expectError: true,
			errContains: "provider region is required",
		},
		{
			name:      "missing token",
			envRegion: "",
			envToken:  "",
			cfg: &AIProvider{
				Model:          "test-model",
				ProviderRegion: "us-east-1",
			},
			expectError: true,
			errContains: "AWS_BEARER_TOKEN_BEDROCK",
		},
		{
			name:      "success with region and token",
			envRegion: "",
			envToken:  "test-key",
			cfg: &AIProvider{
				Model:          "anthropic.claude-3-sonnet",
				ProviderRegion: "us-east-1",
				Temperature:    0.7,
				TopP:           0.9,
			},
			expectError: false,
		},
		{
			name:      "env region overrides config",
			envRegion: "eu-west-1",
			envToken:  "test-key",
			cfg: &AIProvider{
				Model:          "test-model",
				ProviderRegion: "us-east-1",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AWS_DEFAULT_REGION", tt.envRegion)
			t.Setenv("AWS_BEARER_TOKEN_BEDROCK", tt.envToken)

			client := &AmazonBedrockMantleClient{}
			err := client.Configure(tt.cfg)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.cfg.Model, client.model)
			assert.Equal(t, tt.cfg.Temperature, client.temperature)
			assert.Equal(t, tt.cfg.TopP, client.topP)
		})
	}
}

func TestBedrockMantleConfigure_CustomBaseURL(t *testing.T) {
	t.Setenv("AWS_BEARER_TOKEN_BEDROCK", "test-key")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := &AIProvider{
		Model:          "test-model",
		BaseURL:        server.URL,
		ProviderRegion: "us-east-1",
	}
	client := &AmazonBedrockMantleClient{}
	err := client.Configure(cfg)
	assert.NoError(t, err)
}

func TestBedrockMantleGetCompletion_Success(t *testing.T) {
	t.Setenv("AWS_BEARER_TOKEN_BEDROCK", "test-key")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"id":      "chatcmpl-123",
			"object":  "chat.completion",
			"choices": []map[string]interface{}{{"index": 0, "message": map[string]string{"role": "assistant", "content": "mock response"}, "finish_reason": "stop"}},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &AIProvider{
		Model:          "test-model",
		BaseURL:        server.URL,
		ProviderRegion: "us-east-1",
	}
	client := &AmazonBedrockMantleClient{}
	err := client.Configure(cfg)
	assert.NoError(t, err)

	result, err := client.GetCompletion(context.Background(), "hello")
	assert.NoError(t, err)
	assert.Equal(t, "mock response", result)
}

func TestBedrockMantleGetCompletion_Error(t *testing.T) {
	t.Setenv("AWS_BEARER_TOKEN_BEDROCK", "test-key")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{"message": "model error", "type": "server_error"},
		})
	}))
	defer server.Close()

	cfg := &AIProvider{
		Model:          "test-model",
		BaseURL:        server.URL,
		ProviderRegion: "us-east-1",
	}
	client := &AmazonBedrockMantleClient{}
	err := client.Configure(cfg)
	assert.NoError(t, err)

	_, err = client.GetCompletion(context.Background(), "hello")
	assert.Error(t, err)
}
