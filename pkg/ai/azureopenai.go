package ai

import (
	"context"
	"errors"
	"fmt"
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

	defaultConfig.AzureModelMapperFunc = func(model string) string {
		// If you use a deployment name different from the model name, you can customize the AzureModelMapperFunc function
		azureModelMapping := map[string]string{
			model: engine,
		}
		return azureModelMapping[model]

	}

	customHeaders := config.GetCustomHeaders()
	fmt.Println("azureopenai.go Custom Headers:", customHeaders)
	if len(customHeaders) > 0 {
		fmt.Println("azureopenai.go ( customHeaders > 0 ) Custom Headers:", customHeaders)
		transport := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}
		defaultConfig.HTTPClient = &http.Client{
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
			Transport: transport,
		}
	}
	if orgId != "" {
		defaultConfig.OrgID = orgId
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
	return nil
}

func (c *AzureAIClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	// Create a completion request
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: c.temperature,
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func (c *AzureAIClient) GetName() string {
	return azureAIClientName
}
