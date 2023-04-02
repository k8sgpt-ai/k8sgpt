package ai

import (
	"context"
	"errors"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai/openai"
)

var AIProviderMap = map[string]IAI{
	"openai": &openai.OpenAIClient{},
}

type IAI interface {
	Configure(token string, language string) error
	GetCompletion(ctx context.Context, prompt string) (string, error)
	Parse(text string, prompt []string, nocache bool) (string, error)
}

func NewAIClient(provider string) (IAI, error) {
	ai, ok := AIProviderMap[provider]
	if !ok {
		return nil, errors.New("AI provider not found")
	}
	return ai, nil
}
