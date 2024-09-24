package ai

import (
	"context"
	"errors"
	"fmt"

	wx "github.com/IBM/watsonx-go/pkg/models"
)

const ibmWatsonxAIClientName = "ibmwatsonxai"

type IBMWatsonxAIClient struct {
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
	maxTokens      = 2048
)

func (c *IBMWatsonxAIClient) Configure(config IAIConfig) error {
	if config.GetModel() == "" {
		c.model = modelMetallama
	} else {
		c.model = config.GetModel()
	}
	if config.GetMaxTokens() == 0 {
		c.maxNewTokens = maxTokens
	} else {
		c.maxNewTokens = config.GetMaxTokens()
	}
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()
	c.topK = config.GetTopK()

	apiKey := config.GetPassword()
	if apiKey == "" {
		return errors.New("No watsonx API key provided")
	}

	projectId := config.GetProviderId()
	if projectId == "" {
		return errors.New("No watsonx project ID provided")
	}

	client, err := wx.NewClient(
		wx.WithWatsonxAPIKey(apiKey),
		wx.WithWatsonxProjectID(projectId),
	)
	if err != nil {
		return fmt.Errorf("Failed to create client for testing. Error: %v", err)
	}
	c.client = client

	return nil
}

func (c *IBMWatsonxAIClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
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

func (c *IBMWatsonxAIClient) GetName() string {
	return ibmWatsonxAIClientName
}
