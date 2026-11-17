package ai

import (
	"fmt"
	"strings"
)

// checkCompletionContent returns an error when a provider answered with blank
// content, so callers do not treat an empty completion as a successful analysis.
func checkCompletionContent(content, finishReason string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("received an empty completion from the AI provider (finish reason: %s)", finishReason)
	}
	return nil
}
