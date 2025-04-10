package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBedrockModelConfig(t *testing.T) {
	client := AmazonBedRockClient{}

	foundModel, err := client.getModelFromString("arn:aws:bedrock:us-east-1:*:inference-policy/anthropic.claude-3-5-sonnet-20240620-v1:0")
	assert.Nil(t, err, "Error should be nil")
	assert.Equal(t, foundModel.Config.MaxTokens, 100)
	assert.Equal(t, foundModel.Config.Temperature, float32(0.5))
	assert.Equal(t, foundModel.Config.TopP, float32(0.9))
	assert.Equal(t, foundModel.Config.ModelName, "anthropic.claude-3-5-sonnet-20240620-v1:0")
}
