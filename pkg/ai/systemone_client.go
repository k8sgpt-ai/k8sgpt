package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

const systemOneClientName = "systemone"

// SystemOneClient is a purely generic provider that allows K8sGPT to integrate
// with ANY System One model dynamically (e.g. Laya, Jev, or future categorical models).
type SystemOneClient struct {
	nopCloser
	config IAIConfig
}

func (c *SystemOneClient) Configure(config IAIConfig) error {
	c.config = config
	return nil
}

func (c *SystemOneClient) GetName() string {
	return systemOneClientName
}

// GetCompletion is a stub because System One models do not stream conversational text.
func (c *SystemOneClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	return "", errors.New("system one models are deterministic classifiers and do not support autoregressive text generation, use ClassifyAction instead")
}

// ClassifyAction delegates to whichever System One model the user configured in the CLI.
func (c *SystemOneClient) ClassifyAction(ctx context.Context, prompt string, categories []string) (ActionClassification, error) {
	model := c.config.GetModel()
	if model == "" {
		return ActionClassification{}, errors.New("a specific model (e.g. 'laya', 'jev', 'openai-s1') must be specified for the systemone provider")
	}

	baseURL := c.config.GetBaseURL()
	if baseURL == "" {
		return ActionClassification{}, errors.New("base URL must be configured for systemone provider")
	}

	type classificationRequest struct {
		Input      string   `json:"input"`
		Categories []string `json:"categories"`
	}

	reqBody := classificationRequest{
		Input:      prompt,
		Categories: categories,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return ActionClassification{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/classify", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return ActionClassification{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return ActionClassification{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return ActionClassification{}, errors.New("failed to get classification from system one API")
	}

	var result ActionClassification
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ActionClassification{}, err
	}

	return result, nil
}
