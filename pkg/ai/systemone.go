package ai

import "context"

// ActionClassification represents a deterministic classification from a System One model
type ActionClassification struct {
	Decision        string  `json:"decision"`         // e.g. "False Positive", "Actionable", "Security Risk"
	Confidence      float32 `json:"confidence"`       // 0.0 to 1.0 confidence score
	SuggestedAction string  `json:"suggested_action"` // Optional deterministic remediation action (e.g. "RestartPod")
}

// ISystemOne extends the standard AI interface to support non-generative, typed decisions.
// This interface is mathematically bounded (O(1) generation complexity) ensuring it never hangs or loops.
type ISystemOne interface {
	IAI // Inherits GetName(), Configure(), Close(), GetCompletion()

	// ClassifyAction evaluates the prompt against a strictly defined set of categories.
	ClassifyAction(ctx context.Context, prompt string, categories []string) (ActionClassification, error)
}
