package ai

import "context"

type IAI interface {
	Configure(token string, model string, language string) error
	GetCompletion(ctx context.Context, prompt string) (string, error)
}
