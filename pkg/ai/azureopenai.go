package ai

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/sashabaranov/go-openai"
)

const azureAIClientName = "azureopenai"

type AzureAIClient struct {
	nopCloser

	client      *openai.Client
	model       string
	temperature float32
	topP        float32
	maxTokens   int
	// organizationId string
}

type customHeaderRoundTripper struct {
	headers []http.Header
	rt      http.RoundTripper
}

func (c *customHeaderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for _, header := range c.headers {
		for key, values := range header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}
	return c.rt.RoundTrip(req)
}

func (c *AzureAIClient) Configure(config IAIConfig) error {
	token := config.GetPassword()
	baseURL := config.GetBaseURL()
	engine := config.GetEngine()
	proxyEndpoint := config.GetProxyEndpoint()
	defaultConfig := openai.DefaultAzureConfig(token, baseURL)
	orgId := config.GetOrganizationId()
	azureAPIType := config.GetAzureAPIType()
	azureAPIVersion := config.GetAzureAPIVersion()

	if engine != "" {
		defaultConfig.AzureModelMapperFunc = func(model string) string {
			return engine
		}
	}

	customHeaders := config.GetCustomHeaders()
	if len(customHeaders) > 0 {
		transport := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}
		defaultConfig.HTTPClient = &http.Client{
			Timeout: defaultHTTPTimeout,
			Transport: &customHeaderRoundTripper{
				headers: customHeaders,
				rt:      transport,
			},
		}
	} else if proxyEndpoint != "" {
		proxyUrl, err := url.Parse(proxyEndpoint)
		if err != nil {
			return err
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}

		defaultConfig.HTTPClient = &http.Client{
			Timeout:   defaultHTTPTimeout,
			Transport: transport,
		}
	}
	if orgId != "" {
		defaultConfig.OrgID = orgId
	}
	if azureAPIVersion != "" {
		defaultConfig.APIVersion = azureAPIVersion
	}

	switch azureAPIType {
	case string(openai.APITypeAzure):
		defaultConfig.APIType = openai.APITypeAzure
	case string(openai.APITypeAzureAD):
		defaultConfig.APIType = openai.APITypeAzureAD
	case string(openai.APITypeCloudflareAzure):
		defaultConfig.APIType = openai.APITypeCloudflareAzure
	}

	client := openai.NewClientWithConfig(defaultConfig)
	if client == nil {
		return errors.New("error creating Azure OpenAI client")
	}
	c.client = client
	c.model = config.GetModel()
	c.temperature = config.GetTemperature()
	c.topP = config.GetTopP()
	// Left unset (omitted from the request) unless configured, so deployments that
	// reject max_tokens keep working with the default settings.
	c.maxTokens = config.GetMaxTokens()
	return nil
}

func (c *AzureAIClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	// Create a completion request
	req := openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: c.temperature,
		TopP:        c.topP,
		MaxTokens:   c.maxTokens,
	}
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil && rejectsSamplingParams(err) {
		// Reasoning deployments only accept the default sampling values: retry
		// once without them, as the openai backend does.
		req.Temperature, req.TopP = 0, 0
		resp, err = c.client.CreateChatCompletion(ctx, req)
	}
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("no completion choices returned")
	}
	return resp.Choices[0].Message.Content, nil
}

func (c *AzureAIClient) GetName() string {
	return azureAIClientName
}
