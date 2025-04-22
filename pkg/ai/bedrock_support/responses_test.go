package bedrock_support

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCohereResponse_ParseResponse(t *testing.T) {
	response := &CohereResponse{}
	rawResponse := []byte(`{"completion": "Test completion", "stop_reason": "max_tokens"}`)

	result, err := response.ParseResponse(rawResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Test completion", result)

	invalidResponse := []byte(`{"completion": "Test completion", "invalid_json":]`)
	_, err = response.ParseResponse(invalidResponse)
	assert.Error(t, err)
}

func TestAI21Response_ParseResponse(t *testing.T) {
	response := &AI21Response{}
	rawResponse := []byte(`{"completions": [{"data": {"text": "AI21 test"}}], "id": "123"}`)

	result, err := response.ParseResponse(rawResponse)
	assert.NoError(t, err)
	assert.Equal(t, "AI21 test", result)

	invalidResponse := []byte(`{"completions": [{"data": {"text": "AI21 test"}}, "invalid_json":]`)
	_, err = response.ParseResponse(invalidResponse)
	assert.Error(t, err)
}

func TestAmazonResponse_ParseResponse(t *testing.T) {
	response := &AmazonResponse{}
	rawResponse := []byte(`{"inputTextTokenCount": 10, "results": [{"tokenCount": 20, "outputText": "Amazon test", "completionReason": "stop"}]}`)

	result, err := response.ParseResponse(rawResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Amazon test", result)

	invalidResponse := []byte(`{"inputTextTokenCount": 10, "results": [{"tokenCount": 20, "outputText": "Amazon test", "invalid_json":]`)
	_, err = response.ParseResponse(invalidResponse)
	assert.Error(t, err)
}

func TestNovaResponse_ParseResponse(t *testing.T) {
	response := &NovaResponse{}
	rawResponse := []byte(`{"output": {"message": {"content": [{"text": "Nova test"}]}}, "stopReason": "stop", "usage": {"inputTokens": 10, "outputTokens": 20, "totalTokens": 30, "cacheReadInputTokenCount": 5}}`)

	result, err := response.ParseResponse(rawResponse)
	assert.NoError(t, err)
	assert.Equal(t, "Nova test", result)

	rawResponseEmptyContent := []byte(`{"output": {"message": {"content": []}}, "stopReason": "stop", "usage": {"inputTokens": 10, "outputTokens": 20, "totalTokens": 30, "cacheReadInputTokenCount": 5}}`)

	resultEmptyContent, errEmptyContent := response.ParseResponse(rawResponseEmptyContent)
	assert.NoError(t, errEmptyContent)
	assert.Equal(t, "", resultEmptyContent)

	invalidResponse := []byte(`{"output": {"message": {"content": [{"text": "Nova test"}}, "invalid_json":]`)
	_, err = response.ParseResponse(invalidResponse)
	assert.Error(t, err)
}
