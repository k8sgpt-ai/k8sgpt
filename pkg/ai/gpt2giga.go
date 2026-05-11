package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const (
	gpt2gigaClientName = "gpt2giga"
	baseURL            = "http://localhost:8090"
)

type Gpt2GigaClient struct {
	nopCloser

	client      *http.Client
	baseURL     string
	model       string
	temperature float32
	topP        float32
	// organizationId string
}

func (c *Gpt2GigaClient) Configure(config IAIConfig) error {
	c.baseURL = config.GetBaseURL()
	if c.baseURL == "" {
		c.baseURL = baseURL
	}

	c.client = &http.Client{}
	if proxyEndpoint := config.GetProxyEndpoint(); proxyEndpoint != "" {
		proxyUrl, err := url.Parse(proxyEndpoint)
		if err != nil {
			return err
		}
		c.client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
	}

	c.model = config.GetModel()
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()

	return nil
}

func (c *Gpt2GigaClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	// Making a request in the OpenAI format
	requestBody := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": c.temperature,
		"top_p":       c.topP,
	}

	jsonBody, err := json.Marshal(requestBody)

	if err != nil {
		return "", fmt.Errorf("error marshaling request: %w", err)
	}

	// Sending a request
	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Response processing
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", errors.New("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}

func (c *Gpt2GigaClient) GetName() string {
	return gpt2gigaClientName
}
