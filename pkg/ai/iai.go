/*
Copyright 2023 The K8sGPT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ai

import (
	"context"
	"net/http"
)

var (
	clients = []IAI{
		&OpenAIClient{},
		&AzureAIClient{},
		&LocalAIClient{},
		&OllamaClient{},
		&NoOpAIClient{},
		&CohereClient{},
		&AmazonBedRockClient{},
		&SageMakerAIClient{},
		&GoogleGenAIClient{},
		&HuggingfaceClient{},
		&GoogleVertexAIClient{},
		&OCIGenAIClient{},
		&CustomRestClient{},
		&IBMWatsonxAIClient{},
		&LiteLLMClient{},
	}
	Backends = []string{
		openAIClientName,
		localAIClientName,
		ollamaClientName,
		azureAIClientName,
		cohereAIClientName,
		amazonbedrockAIClientName,
		amazonsagemakerAIClientName,
		googleAIClientName,
		noopAIClientName,
		huggingfaceAIClientName,
		googleVertexAIClientName,
		ociClientName,
		CustomRestClientName,
		ibmWatsonxAIClientName,
		liteLLMClientName,
	}
)

// IAI is an interface all clients (representing backends) share.
type IAI interface {
	// Configure sets up client for given configuration. This is expected to be
	// executed once per client life-time (e.g. analysis CLI command invocation).
	Configure(config IAIConfig) error
	// GetCompletion generates text based on prompt.
	GetCompletion(ctx context.Context, prompt string) (string, error)
	// GetName returns name of the backend/client.
	GetName() string
	// Close cleans all the resources. No other methods should be used on the
	// objects after this method is invoked.
	Close()
}

type nopCloser struct{}

func (nopCloser) Close() {}

type IAIConfig interface {
	GetPassword() string
	GetModel() string
	GetBaseURL() string
	GetProxyEndpoint() string
	GetEndpointName() string
	GetEngine() string
	GetTemperature() float32
	GetProviderRegion() string
	GetTopP() float32
	GetTopK() int32
	GetMaxTokens() int
	GetProviderId() string
	GetCompartmentId() string
	GetOrganizationId() string
	GetCustomHeaders() []http.Header
}

func NewClient(provider string) IAI {
	for _, c := range clients {
		if provider == c.GetName() {
			return c
		}
	}
	// default client
	return &OpenAIClient{}
}

type AIConfiguration struct {
	Providers       []AIProvider `mapstructure:"providers"`
	DefaultProvider string       `mapstructure:"defaultprovider"`
}

type AIProvider struct {
	Name           string        `mapstructure:"name"`
	Model          string        `mapstructure:"model"`
	Password       string        `mapstructure:"password" yaml:"password,omitempty"`
	BaseURL        string        `mapstructure:"baseurl" yaml:"baseurl,omitempty"`
	ProxyEndpoint  string        `mapstructure:"proxyEndpoint" yaml:"proxyEndpoint,omitempty"`
	ProxyPort      string        `mapstructure:"proxyPort" yaml:"proxyPort,omitempty"`
	EndpointName   string        `mapstructure:"endpointname" yaml:"endpointname,omitempty"`
	Engine         string        `mapstructure:"engine" yaml:"engine,omitempty"`
	Temperature    float32       `mapstructure:"temperature" yaml:"temperature,omitempty"`
	ProviderRegion string        `mapstructure:"providerregion" yaml:"providerregion,omitempty"`
	ProviderId     string        `mapstructure:"providerid" yaml:"providerid,omitempty"`
	CompartmentId  string        `mapstructure:"compartmentid" yaml:"compartmentid,omitempty"`
	TopP           float32       `mapstructure:"topp" yaml:"topp,omitempty"`
	TopK           int32         `mapstructure:"topk" yaml:"topk,omitempty"`
	MaxTokens      int           `mapstructure:"maxtokens" yaml:"maxtokens,omitempty"`
	OrganizationId string        `mapstructure:"organizationid" yaml:"organizationid,omitempty"`
	CustomHeaders  []http.Header `mapstructure:"customHeaders"`
}

func (p *AIProvider) GetBaseURL() string {
	return p.BaseURL
}

func (p *AIProvider) GetProxyEndpoint() string {
	return p.ProxyEndpoint
}

func (p *AIProvider) GetEndpointName() string {
	return p.EndpointName
}

func (p *AIProvider) GetTopP() float32 {
	return p.TopP
}

func (p *AIProvider) GetTopK() int32 {
	return p.TopK
}

func (p *AIProvider) GetMaxTokens() int {
	return p.MaxTokens
}

func (p *AIProvider) GetPassword() string {
	return p.Password
}

func (p *AIProvider) GetModel() string {
	return p.Model
}

func (p *AIProvider) GetEngine() string {
	return p.Engine
}
func (p *AIProvider) GetTemperature() float32 {
	return p.Temperature
}

func (p *AIProvider) GetProviderRegion() string {
	return p.ProviderRegion
}

func (p *AIProvider) GetProviderId() string {
	return p.ProviderId
}

func (p *AIProvider) GetCompartmentId() string {
	return p.CompartmentId
}

func (p *AIProvider) GetOrganizationId() string {
	return p.OrganizationId
}

func (p *AIProvider) GetCustomHeaders() []http.Header {
	return p.CustomHeaders
}

var passwordlessProviders = []string{"localai", "ollama", "amazonsagemaker", "amazonbedrock", "googlevertexai", "oci", "customrest", "litellm"}

func NeedPassword(backend string) bool {
	for _, b := range passwordlessProviders {
		if b == backend {
			return false
		}
	}
	return true
}
