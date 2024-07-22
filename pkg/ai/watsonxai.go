package ai

import (
	"context"
	"errors"
	"fmt"
	"os"

	wx "github.com/IBM/watsonx-go/pkg/models"
)

const watsonxAIClientName = "watsonxai"

type WatsonxAIClient struct {
	nopCloser

	client       *wx.Client
	model        string
	temperature  float32
	topP         float32
	topK         int32
	maxNewTokens int
}

const (
	modelMetallama = "ibm/granite-13b-chat-v2"
)

func (c *WatsonxAIClient) Configure(config IAIConfig) error {
	if config.GetModel() == "" {
		c.model = config.GetModel()
	} else {
		c.model = modelMetallama
	}
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()
	c.topK = config.GetTopK()
	c.maxNewTokens = config.GetMaxTokens()

	// WatsonxAPIKeyEnvVarName    = "WATSONX_API_KEY"
	// WatsonxProjectIDEnvVarName = "WATSONX_PROJECT_ID"
	apiKey, projectID := os.Getenv(wx.WatsonxAPIKeyEnvVarName), os.Getenv(wx.WatsonxProjectIDEnvVarName)

	if apiKey == "" {
		return errors.New("No watsonx API key provided")
	}
	if projectID == "" {
		return errors.New("No watsonx project ID provided")
	}

	client, err := wx.NewClient(
		wx.WithWatsonxAPIKey(apiKey),
		wx.WithWatsonxProjectID(projectID),
	)
	if err != nil {
		return fmt.Errorf("Failed to create client for testing. Error: %v", err)
	}
	c.client = client

	return nil
}

func (c *WatsonxAIClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	result, err := c.client.GenerateText(
		c.model,
		prompt,
		wx.WithTemperature((float64)(c.temperature)),
		wx.WithTopP((float64)(c.topP)),
		wx.WithTopK((uint)(c.topK)),
		wx.WithMaxNewTokens((uint)(c.maxNewTokens)),
	)
	if err != nil {
		return "", fmt.Errorf("Expected no error, but got an error: %v", err)
	}
	if result.Text == "" {
		return "", errors.New("Expected a result, but got an empty string")
	}

	return result.Text, nil
}

func (c *WatsonxAIClient) GetName() string {
	return watsonxAIClientName
}
