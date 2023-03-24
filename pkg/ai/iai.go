package ai

import "context"

type IAI interface {
	Configure(token string) error
	GetCompletion(ctx context.Context, prompt string) (string, error)
}
