package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/viper"
)

const (
	openAIResponsesClientName    = "openairesponses"
	defaultOpenAIResponsesURL    = "https://api.openai.com/v1/responses"
	defaultOpenAIResponsesModel  = "gpt-5-nano"
	defaultOpenAIResponsesTokens = 2048
)

type OpenAIResponsesClient struct {
	nopCloser

	client         *http.Client
	endpointURL    string
	token          string
	model          string
	temperature    float32
	topP           float32
	maxOutputToken int
	organizationID string
}

type openAIResponsesRequest struct {
	Model           string                    `json:"model"`
	Input           string                    `json:"input"`
	MaxOutputTokens int                       `json:"max_output_tokens,omitempty"`
	Temperature     *float32                  `json:"temperature,omitempty"`
	TopP            *float32                  `json:"top_p,omitempty"`
	Reasoning       *openAIResponsesReasoning `json:"reasoning,omitempty"`
	Text            *openAIResponsesText      `json:"text,omitempty"`
}

type openAIResponsesResponse struct {
	Output     []openAIResponsesOutputItem `json:"output"`
	OutputText string                    `json:"output_text,omitempty"`
	Error      *openAIResponsesError      `json:"error,omitempty"`
}

type openAIResponsesOutputItem struct {
	Type    string                     `json:"type"`
	Content []openAIResponsesTextChunk `json:"content,omitempty"`
}

type openAIResponsesTextChunk struct {
	Type    string `json:"type"`
	Text    string `json:"text,omitempty"`
	Refusal string `json:"refusal,omitempty"`
}

type openAIResponsesError struct {
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
	Code    string `json:"code,omitempty"`
}

type openAIResponsesReasoning struct {
	Effort  string `json:"effort,omitempty"`
	Summary string `json:"summary,omitempty"`
}

type openAIResponsesText struct {
	Format    openAIResponsesTextFormat `json:"format,omitempty"`
	Verbosity string                    `json:"verbosity,omitempty"`
}

type openAIResponsesTextFormat struct {
	Type string `json:"type"`
}

type openAIErrorResponse struct {
	Error openAIResponsesError `json:"error"`
}

func (c *OpenAIResponsesClient) Configure(config IAIConfig) error {
	c.token = config.GetPassword()
	c.model = config.GetModel()
	if c.model == "" {
		c.model = defaultOpenAIResponsesModel
	}
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()
	c.maxOutputToken = config.GetMaxTokens()
	if c.maxOutputToken <= 0 {
		c.maxOutputToken = defaultOpenAIResponsesTokens
	}
	c.organizationID = config.GetOrganizationId()

	endpointURL, err := buildOpenAIResponsesURL(config.GetBaseURL())
	if err != nil {
		return err
	}
	c.endpointURL = endpointURL

	transport := http.DefaultTransport.(*http.Transport).Clone()
	proxyEndpoint := config.GetProxyEndpoint()
	if proxyEndpoint != "" {
		proxyURL, err := url.Parse(proxyEndpoint)
		if err != nil {
			return err
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	customHeaders := config.GetCustomHeaders()
	if len(customHeaders) > 0 {
		c.client = &http.Client{
			Transport: &OpenAIHeaderTransport{
				Origin:  transport,
				Headers: customHeaders,
			},
		}
	} else {
		c.client = &http.Client{Transport: transport}
	}

	return nil
}

func (c *OpenAIResponsesClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	payload := openAIResponsesRequest{
		Model:           c.model,
		Input:           prompt,
		MaxOutputTokens: c.maxOutputToken,
	}

	if supportsOpenAIResponsesSamplingParams(c.model) {
		payload.Temperature = &c.temperature
		payload.TopP = &c.topP
	}
	if requiresOpenAIResponsesReasoningConfig(c.model) {
		payload.Reasoning = &openAIResponsesReasoning{
			Effort: "low",
		}
	}
	payload.Text = &openAIResponsesText{
		Format: openAIResponsesTextFormat{Type: "text"},
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpointURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if c.organizationID != "" {
		request.Header.Set("OpenAI-Organization", c.organizationID)
	}

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
		return "", parseOpenAIError(responseBody, response.StatusCode)
	}

	var parsed openAIResponsesResponse
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return "", err
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", fmt.Errorf("OpenAI error: %s", parsed.Error.Message)
	}

	text := extractOpenAIResponsesText(parsed)
	if text == "" {
		if viper.GetBool("verbose") {
			return "", fmt.Errorf("no output text found in OpenAI response. Raw response: %s", truncateForLog(string(responseBody), 4000))
		}
		return "", errors.New("no output text found in OpenAI response")
	}
	return text, nil
}

func (c *OpenAIResponsesClient) GetName() string {
	return openAIResponsesClientName
}

func buildOpenAIResponsesURL(baseURL string) (string, error) {
	if baseURL == "" {
		return defaultOpenAIResponsesURL, nil
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	path := strings.TrimRight(parsed.Path, "/")
	if strings.HasSuffix(path, "/responses") {
		return parsed.String(), nil
	}

	parsed.Path = path + "/responses"
	return parsed.String(), nil
}

func extractOpenAIResponsesText(response openAIResponsesResponse) string {
	var builder strings.Builder
	for _, item := range response.Output {
		if item.Type != "message" {
			continue
		}
		for _, content := range item.Content {
			switch content.Type {
			case "output_text", "reasoning_text", "summary_text":
				if content.Text == "" {
					continue
				}
				builder.WriteString(content.Text)
			case "refusal":
				if content.Refusal == "" {
					continue
				}
				builder.WriteString(content.Refusal)
			}
		}
	}
	text := strings.TrimSpace(builder.String())
	if text == "" && response.OutputText != "" {
		return strings.TrimSpace(response.OutputText)
	}
	return text
}

func supportsOpenAIResponsesSamplingParams(model string) bool {
	normalized := strings.ToLower(strings.TrimSpace(model))
	if normalized == "" {
		return true
	}
	if strings.HasPrefix(normalized, "gpt-5") {
		return false
	}
	if strings.HasPrefix(normalized, "o1") || strings.HasPrefix(normalized, "o3") || strings.HasPrefix(normalized, "o4") {
		return false
	}
	return true
}

func requiresOpenAIResponsesReasoningConfig(model string) bool {
	normalized := strings.ToLower(strings.TrimSpace(model))
	if normalized == "" {
		return false
	}
	if strings.HasPrefix(normalized, "gpt-5") {
		return true
	}
	if strings.HasPrefix(normalized, "o1") || strings.HasPrefix(normalized, "o3") || strings.HasPrefix(normalized, "o4") {
		return true
	}
	return false
}

func parseOpenAIError(responseBody []byte, statusCode int) error {
	var errResponse openAIErrorResponse
	if err := json.Unmarshal(responseBody, &errResponse); err == nil && errResponse.Error.Message != "" {
		return fmt.Errorf("OpenAI request failed (status %d): %s", statusCode, errResponse.Error.Message)
	}
	return fmt.Errorf("OpenAI request failed (status %d): %s", statusCode, strings.TrimSpace(string(responseBody)))
}

func truncateForLog(value string, maxChars int) string {
	trimmed := strings.TrimSpace(value)
	if maxChars <= 0 || len(trimmed) <= maxChars {
		return trimmed
	}
	return trimmed[:maxChars] + "...(truncated)"
}
