package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const CustomRestClientName = "customrest"

type CustomRestClient struct {
	nopCloser
	client      *http.Client
	base        *url.URL
	token       string
	model       string
	temperature float32
	topP        float32
	topK        int32
}

type CustomRestRequest struct {
	Model string `json:"model"`

	// Prompt is the textual prompt to send to the model.
	Prompt string `json:"prompt"`

	// Options lists model-specific options. For example, temperature can be
	// set through this field, if the model supports it.
	Options map[string]interface{} `json:"options"`
}

type CustomRestResponse struct {
	// Model is the model name that generated the response.
	Model string `json:"model"`

	// CreatedAt is the timestamp of the response.
	CreatedAt time.Time `json:"created_at"`

	// Response is the textual response itself.
	Response string `json:"response"`
}

func (c *CustomRestClient) Configure(config IAIConfig) error {
	baseURL := config.GetBaseURL()
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	c.token = config.GetPassword()
	baseClientURL, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	c.base = baseClientURL

	proxyEndpoint := config.GetProxyEndpoint()
	c.client = http.DefaultClient
	if proxyEndpoint != "" {
		proxyUrl, err := url.Parse(proxyEndpoint)
		if err != nil {
			return err
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}

		c.client = &http.Client{
			Transport: transport,
		}
	}

	c.model = config.GetModel()
	if c.model == "" {
		c.model = defaultModel
	}
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()
	c.topK = config.GetTopK()
	return nil
}

func (c *CustomRestClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	var promptDetail struct {
		Language string `json:"language,omitempty"`
		Message  string `json:"message"`
		Prompt   string `json:"prompt,omitempty"`
	}
	prompt = strings.NewReplacer("\n", "\\n", "\t", "\\t").Replace(prompt)
	if err := json.Unmarshal([]byte(prompt), &promptDetail); err != nil {
		return "", err
	}
	generateRequest := &CustomRestRequest{
		Model:  c.model,
		Prompt: promptDetail.Prompt,
		Options: map[string]interface{}{
			"temperature": c.temperature,
			"top_p":       c.topP,
			"top_k":       c.topK,
			"message":     promptDetail.Message,
			"language":    promptDetail.Language,
		},
	}
	requestBody, err := json.Marshal(generateRequest)
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base.String(), bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	if c.token != "" {
		request.Header.Set("Authorization", "Bearer "+c.token)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/x-ndjson")

	response, err := c.client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("could not read response body: %w", err)
	}

	if response.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("Request Error, StatusCode: %d, ErrorMessage: %s", response.StatusCode, responseBody)
	}

	var result CustomRestResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return "", err
	}
	return result.Response, nil
}

func (c *CustomRestClient) GetName() string {
	return CustomRestClientName
}
