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

// IAIConfig represents the configuration for an AI provider
type IAIConfig interface {
	GetModel() string
	GetProviderRegion() string
	GetTemperature() float32
	GetTopP() float32
	GetMaxTokens() int
	GetConfigName() string // Added to support multiple configurations
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

// AIProvider represents a provider configuration
type AIProvider struct {
	Name           string             `mapstructure:"name" json:"name"`
	Model          string             `mapstructure:"model" json:"model,omitempty"`
	Password       string             `mapstructure:"password" yaml:"password,omitempty" json:"password,omitempty"`
	BaseURL        string             `mapstructure:"baseurl" yaml:"baseurl,omitempty" json:"baseurl,omitempty"`
	ProxyEndpoint  string             `mapstructure:"proxyEndpoint" yaml:"proxyEndpoint,omitempty" json:"proxyEndpoint,omitempty"`
	ProxyPort      string             `mapstructure:"proxyPort" yaml:"proxyPort,omitempty" json:"proxyPort,omitempty"`
	EndpointName   string             `mapstructure:"endpointname" yaml:"endpointname,omitempty" json:"endpointname,omitempty"`
	Engine         string             `mapstructure:"engine" yaml:"engine,omitempty" json:"engine,omitempty"`
	Temperature    float32            `mapstructure:"temperature" yaml:"temperature,omitempty" json:"temperature,omitempty"`
	ProviderRegion string             `mapstructure:"providerregion" yaml:"providerregion,omitempty" json:"providerregion,omitempty"`
	ProviderId     string             `mapstructure:"providerid" yaml:"providerid,omitempty" json:"providerid,omitempty"`
	CompartmentId  string             `mapstructure:"compartmentid" yaml:"compartmentid,omitempty" json:"compartmentid,omitempty"`
	TopP           float32            `mapstructure:"topp" yaml:"topp,omitempty" json:"topp,omitempty"`
	TopK           int32              `mapstructure:"topk" yaml:"topk,omitempty" json:"topk,omitempty"`
	MaxTokens      int                `mapstructure:"maxtokens" yaml:"maxtokens,omitempty" json:"maxtokens,omitempty"`
	OrganizationId string             `mapstructure:"organizationid" yaml:"organizationid,omitempty" json:"organizationid,omitempty"`
	CustomHeaders  []http.Header      `mapstructure:"customHeaders" json:"customHeaders,omitempty"`
	Configs        []AIProviderConfig `mapstructure:"configs" json:"configs,omitempty"`
	DefaultConfig  int                `mapstructure:"defaultConfig" json:"defaultConfig,omitempty"`
}

// AIProviderConfig represents a single configuration for a provider
type AIProviderConfig struct {
	Model          string        `mapstructure:"model" json:"model"`
	ProviderRegion string        `mapstructure:"providerRegion" json:"providerRegion"`
	Temperature    float32       `mapstructure:"temperature" json:"temperature"`
	TopP           float32       `mapstructure:"topP" json:"topP"`
	MaxTokens      int           `mapstructure:"maxTokens" json:"maxTokens"`
	ConfigName     string        `mapstructure:"configName" json:"configName"`
	Password       string        `mapstructure:"password" yaml:"password,omitempty" json:"password,omitempty"`
	BaseURL        string        `mapstructure:"baseurl" yaml:"baseurl,omitempty" json:"baseurl,omitempty"`
	ProxyEndpoint  string        `mapstructure:"proxyEndpoint" yaml:"proxyEndpoint,omitempty" json:"proxyEndpoint,omitempty"`
	EndpointName   string        `mapstructure:"endpointname" yaml:"endpointname,omitempty" json:"endpointname,omitempty"`
	Engine         string        `mapstructure:"engine" yaml:"engine,omitempty" json:"engine,omitempty"`
	ProviderId     string        `mapstructure:"providerid" yaml:"providerid,omitempty" json:"providerid,omitempty"`
	CompartmentId  string        `mapstructure:"compartmentid" yaml:"compartmentid,omitempty" json:"compartmentid,omitempty"`
	TopK           int32         `mapstructure:"topk" yaml:"topk,omitempty" json:"topk,omitempty"`
	OrganizationId string        `mapstructure:"organizationid" yaml:"organizationid,omitempty" json:"organizationid,omitempty"`
	CustomHeaders  []http.Header `mapstructure:"customHeaders" json:"customHeaders,omitempty"`
}

// GetConfigName returns the configuration name
func (p *AIProvider) GetConfigName() string {
	if len(p.Configs) > 0 && p.DefaultConfig >= 0 && p.DefaultConfig < len(p.Configs) {
		return p.Configs[p.DefaultConfig].ConfigName
	}
	return ""
}

// GetModel returns the model name
func (p *AIProvider) GetModel() string {
	if len(p.Configs) > 0 && p.DefaultConfig >= 0 && p.DefaultConfig < len(p.Configs) {
		return p.Configs[p.DefaultConfig].Model
	}
	return p.Model
}

// GetProviderRegion returns the provider region
func (p *AIProvider) GetProviderRegion() string {
	if len(p.Configs) > 0 && p.DefaultConfig >= 0 && p.DefaultConfig < len(p.Configs) {
		return p.Configs[p.DefaultConfig].ProviderRegion
	}
	return p.ProviderRegion
}

// GetTemperature returns the temperature
func (p *AIProvider) GetTemperature() float32 {
	if len(p.Configs) > 0 && p.DefaultConfig >= 0 && p.DefaultConfig < len(p.Configs) {
		return p.Configs[p.DefaultConfig].Temperature
	}
	return p.Temperature
}

// GetTopP returns the top P value
func (p *AIProvider) GetTopP() float32 {
	if len(p.Configs) > 0 && p.DefaultConfig >= 0 && p.DefaultConfig < len(p.Configs) {
		return p.Configs[p.DefaultConfig].TopP
	}
	return p.TopP
}

// GetMaxTokens returns the maximum number of tokens
func (p *AIProvider) GetMaxTokens() int {
	if len(p.Configs) > 0 && p.DefaultConfig >= 0 && p.DefaultConfig < len(p.Configs) {
		return p.Configs[p.DefaultConfig].MaxTokens
	}
	return p.MaxTokens
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

func (p *AIProvider) GetTopK() int32 {
	return p.TopK
}

func (p *AIProvider) GetPassword() string {
	return p.Password
}

func (p *AIProvider) GetEngine() string {
	return p.Engine
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

var passwordlessProviders = []string{"localai", "ollama", "amazonsagemaker", "amazonbedrock", "googlevertexai", "oci", "customrest"}

func NeedPassword(backend string) bool {
	for _, b := range passwordlessProviders {
		if b == backend {
			return false
		}
	}
	return true
}
