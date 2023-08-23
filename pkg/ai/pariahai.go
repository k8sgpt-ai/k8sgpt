package ai

import (
	"context"

	"github.com/k8sgpt-ai/k8sgpt/pkg/cache"
)

// implement IAI interface

type PariahAIClient struct {
}

func (c *PariahAIClient) Configure(config IAIConfig, lang string) error {
	return nil
}

func (c *PariahAIClient) GetCompletion(ctx context.Context, prompt string, promptTmpl string) (string, error) {
	return "", nil
}

func (c *PariahAIClient) Parse(ctx context.Context, prompt []string, cache cache.ICache, promptTmpl string) (string, error) {

	return "", nil
}

func (c *PariahAIClient) GetName() string {

	return "pariahai"
}

func (c *PariahAIClient) GetPassword() string {
	return ""
}
