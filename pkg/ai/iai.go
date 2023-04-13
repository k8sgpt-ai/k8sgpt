package ai

import (
	"context"
)

type IAI interface {
	Configure(token string, model string, language string) error
	GetCompletion(ctx context.Context, prompt string) (string, error)
	Parse(ctx context.Context, prompt []string, nocache bool) (string, error)
	GetName() string
}

func NewClient(provider string) IAI {
	switch provider {
	case "openai":
		return &OpenAIClient{}
	case "noopai":
		return &NoOpAIClient{}
	default:
		return &OpenAIClient{}
	}
}

type AIConfiguration struct {
	Providers []AIProvider `mapstructure:"providers"`
}

type AIProvider struct {
	Name       string `mapstructure:"name"`
	Model      string `mapstructure:"model"`
	Password   string `mapstructure:"password"`
	Passphrase string `mapstructure:"passphrase"`
}
