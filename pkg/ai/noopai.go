package ai

import (
	"context"
	"encoding/base64"
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"strings"
)

type NoOpAIClient struct {
	client   string
	language string
}

func (c *NoOpAIClient) Configure(token string, language string) error {
	c.language = language
	c.client = "I am a noop client with the token " + token
	return nil
}

func (c *NoOpAIClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	// Create a completion request
	response := "I am a noop response to the prompt " + prompt
	return response, nil
}

func (a *NoOpAIClient) Parse(ctx context.Context, prompt []string, nocache bool) (string, error) {
	// parse the text with the AI backend
	inputKey := strings.Join(prompt, " ")
	// Check for cached data
	sEnc := base64.StdEncoding.EncodeToString([]byte(inputKey))

	response, err := a.GetCompletion(ctx, inputKey)
	if err != nil {
		color.Red("error getting completion: %v", err)
		return "", err
	}

	if !viper.IsSet(sEnc) {
		viper.Set(sEnc, base64.StdEncoding.EncodeToString([]byte(response)))
		if err := viper.WriteConfig(); err != nil {
			color.Red("error writing config: %v", err)
			return "", nil
		}
	}
	return response, nil
}

func (a *NoOpAIClient) GetName() string {
	return "noopai"
}
