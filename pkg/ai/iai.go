package ai

import (
	"context"
)

type IAI interface {
	Configure(token string, language string) error
	GetCompletion(ctx context.Context, prompt string) (string, error)
	Parse(ctx context.Context, prompt []string, nocache bool) (string, error)
}

func NewClient(provider string) IAI {
	switch provider {
	case "openai":
		return &OpenAIClient{}
	default:
		return &OpenAIClient{}
	}
}
