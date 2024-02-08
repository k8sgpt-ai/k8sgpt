package ai

import (
	"context"
	"github.com/hupe1980/go-huggingface"
	"k8s.io/utils/ptr"
)

const huggingfaceAIClientName = "huggingface"

type HuggingfaceClient struct {
	nopCloser

	client      *huggingface.InferenceClient
	model       string
	topP        float32
	temperature float32
	maxTokens   int
}

func (c *HuggingfaceClient) Configure(config IAIConfig) error {
	token := config.GetPassword()

	client := huggingface.NewInferenceClient(token)

	c.client = client
	c.model = config.GetModel()
	c.topP = config.GetTopP()
	c.temperature = config.GetTemperature()
	if config.GetMaxTokens() > 500 {
		c.maxTokens = 500
	} else {
		c.maxTokens = config.GetMaxTokens()
	}
	return nil
}

func (c *HuggingfaceClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	resp, err := c.client.Conversational(ctx, &huggingface.ConversationalRequest{
		Inputs: huggingface.ConverstationalInputs{
			Text: prompt,
		},
		Model: c.model,
		Parameters: huggingface.ConversationalParameters{
			TopP:        ptr.To[float64](float64(c.topP)),
			Temperature: ptr.To[float64](float64(c.temperature)),
			MaxLength:   &c.maxTokens,
		},
		Options: huggingface.Options{
			WaitForModel: ptr.To[bool](true),
		},
	})
	if err != nil {
		return "", err
	}
	return resp.GeneratedText, nil
}

func (c *HuggingfaceClient) GetName() string { return huggingfaceAIClientName }
