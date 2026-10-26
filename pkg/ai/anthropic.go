package ai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	anthropicClientName       = "anthropic"
	anthropicDefaultBaseURL   = "https://api.anthropic.com"
	anthropicMessagesPath     = "/v1/messages"
	anthropicDefaultModel     = "claude-3-5-sonnet-latest"
	anthropicDefaultMaxTokens = 2048
)

type AnthropicClient struct {
	nopCloser

	client        anthropic.Client
	token         string
	model         string
	temperature   float32
	topP          float32
	topK          int32
	maxTokens     int
	stopSequences []string
	customHeaders []http.Header
}

func (c *AnthropicClient) Configure(config IAIConfig) error {
	baseURL, err := anthropicBaseURL(config.GetBaseURL())
	if err != nil {
		return err
	}

	var opts []option.RequestOption
	opts = append(opts,
		option.WithoutEnvironmentDefaults(),
		option.WithBaseURL(baseURL),
	)
	if token := config.GetPassword(); token != "" {
		opts = append(opts, option.WithAPIKey(token))
	}

	httpClient := &http.Client{Timeout: defaultHTTPTimeout}
	proxyEndpoint := config.GetProxyEndpoint()
	if proxyEndpoint != "" {
		proxyURL, err := url.Parse(proxyEndpoint)
		if err != nil {
			return err
		}
		httpClient.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	}
	opts = append(opts, option.WithHTTPClient(httpClient))

	model := config.GetModel()
	if model == "" {
		model = anthropicDefaultModel
	}
	maxTokens := config.GetMaxTokens()
	if maxTokens <= 0 {
		maxTokens = anthropicDefaultMaxTokens
	}

	for key, values := range mergeCustomHeaders(config.GetCustomHeaders()) {
		if len(values) == 0 {
			continue
		}
		opts = append(opts, option.WithHeaderDel(key), option.WithHeader(key, values[0]))
		for _, value := range values[1:] {
			opts = append(opts, option.WithHeaderAdd(key, value))
		}
	}

	c.client = anthropic.NewClient(opts...)
	c.token = config.GetPassword()
	c.model = model
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()
	c.topK = config.GetTopK()
	c.maxTokens = maxTokens
	c.stopSequences = config.GetStopSequences()
	c.customHeaders = config.GetCustomHeaders()
	return nil
}

func (c *AnthropicClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	params := anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: int64(c.maxTokens),
		Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(prompt))},
	}
	// Anthropic models only support temperature OR top_p, not both. Prefer temperature.
	if c.topP > 0 && c.temperature == 0 {
		params.TopP = anthropic.Float(float64(c.topP))
	} else {
		params.Temperature = anthropic.Float(float64(c.temperature))
	}
	if c.topK > 0 {
		params.TopK = anthropic.Int(int64(c.topK))
	}
	if len(c.stopSequences) > 0 {
		params.StopSequences = c.stopSequences
	}

	var textBlocks []string
	message, err := c.client.Messages.New(ctx, params)
	if err != nil {
		return "", err
	}
	for _, content := range message.Content {
		if content.Type == "text" {
			text := content.AsText().Text
			if text != "" {
				textBlocks = append(textBlocks, text)
			}
		}
	}
	if len(textBlocks) == 0 {
		return "", errors.New("anthropic response did not include any text content")
	}
	return strings.Join(textBlocks, "\n"), nil
}

func (c *AnthropicClient) GetName() string {
	return anthropicClientName
}

func anthropicBaseURL(rawBaseURL string) (string, error) {
	if rawBaseURL == "" {
		rawBaseURL = anthropicDefaultBaseURL
	}

	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return "", err
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return "", fmt.Errorf("invalid anthropic base URL %q", rawBaseURL)
	}

	switch {
	case strings.HasSuffix(baseURL.Path, anthropicMessagesPath), strings.HasSuffix(baseURL.Path, "/messages"):
		baseURL.Path = strings.TrimSuffix(baseURL.Path, anthropicMessagesPath)
		baseURL.Path = strings.TrimSuffix(baseURL.Path, "/messages")
		if baseURL.Path == "" {
			baseURL.Path = "/"
		}
		return strings.TrimRight(baseURL.String(), "/"), nil
	default:
		baseURL.Path = path.Clean(baseURL.Path)
		if baseURL.Path == "." {
			baseURL.Path = ""
		}
		return strings.TrimRight(baseURL.String(), "/"), nil
	}
}

func mergeCustomHeaders(headers []http.Header) http.Header {
	merged := http.Header{}
	for _, header := range headers {
		for key, values := range header {
			copiedValues := make([]string, len(values))
			copy(copiedValues, values)
			merged[key] = append(merged[key], copiedValues...)
		}
	}
	return merged
}
