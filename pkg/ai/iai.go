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
)

var (
	clients = []IAI{
		&OpenAIClient{},
		&AzureAIClient{},
		&LocalAIClient{},
		&NoOpAIClient{},
		&CohereClient{},
		&AmazonBedRockClient{},
		&SageMakerAIClient{},
		&GoogleGenAIClient{},
		&HuggingfaceClient{},
		&GoogleVertexAIClient{},
	}
	Backends = []string{
		openAIClientName,
		localAIClientName,
		azureAIClientName,
		cohereAIClientName,
		amazonbedrockAIClientName,
		amazonsagemakerAIClientName,
		googleAIClientName,
		noopAIClientName,
		huggingfaceAIClientName,
		googleVertexAIClientName,
	}
)

// IAI is an interface all clients (representing backends) share.
type IAI interface {
	// Configure sets up client for given configuration. This is expected to be
	// executed once per client life-time (e.g. analysis CLI command invocation).
	Configure(config IAIConfig, index int) error
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
	GetPassword(index int) string
	GetModel(index int) string
	GetBaseURL(index int) string
	GetProxyEndpoint(index int) string
	GetEndpointName(index int) string
	GetEngine(index int) string
	GetTemperature(index int) float32
	GetProviderRegion(index int) string
	GetTopP(index int) float32
	GetMaxTokens(index int) int
	GetProviderId(index int) string
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
	Backend       string
	Configs       []AIProviderConfig
	DefaultConfig int
}

type AIProviderConfig struct {
	Name           string  `mapstructure:"name"`
	Model          string  `mapstructure:"model"`
	Password       string  `mapstructure:"password" yaml:"password,omitempty"`
	BaseURL        string  `mapstructure:"baseurl" yaml:"baseurl,omitempty"`
	ProxyEndpoint  string  `mapstructure:"proxyEndpoint" yaml:"proxyEndpoint,omitempty"`
	ProxyPort      string  `mapstructure:"proxyPort" yaml:"proxyPort,omitempty"`
	EndpointName   string  `mapstructure:"endpointname" yaml:"endpointname,omitempty"`
	Engine         string  `mapstructure:"engine" yaml:"engine,omitempty"`
	Temperature    float32 `mapstructure:"temperature" yaml:"temperature,omitempty"`
	ProviderRegion string  `mapstructure:"providerregion" yaml:"providerregion,omitempty"`
	ProviderId     string  `mapstructure:"providerid" yaml:"providerid,omitempty"`
	TopP           float32 `mapstructure:"topp" yaml:"topp,omitempty"`
	MaxTokens      int     `mapstructure:"maxtokens" yaml:"maxtokens,omitempty"`
}

func (p *AIProvider) GetBaseURL(index int) string {
	return p.Configs[index].BaseURL
}

func (p *AIProvider) GetProxyEndpoint(index int) string {
	return p.Configs[index].ProxyEndpoint
}

func (p *AIProvider) GetEndpointName(index int) string {
	return p.Configs[index].EndpointName
}

func (p *AIProvider) GetTopP(index int) float32 {
	return p.Configs[index].TopP
}

func (p *AIProvider) GetMaxTokens(index int) int {
	return p.Configs[index].MaxTokens
}

func (p *AIProvider) GetPassword(index int) string {
	return p.Configs[index].Password
}

func (p *AIProvider) GetModel(index int) string {
	return p.Configs[index].Model
}

func (p *AIProvider) GetEngine(index int) string {
	return p.Configs[index].Engine
}
func (p *AIProvider) GetTemperature(index int) float32 {
	return p.Configs[index].Temperature
}

func (p *AIProvider) GetProviderRegion(index int) string {
	return p.Configs[index].ProviderRegion
}

func (p *AIProvider) GetProviderId(index int) string {
	return p.Configs[index].ProviderId
}

var passwordlessProviders = []string{"localai", "amazonsagemaker", "amazonbedrock", "googlevertexai"}

func NeedPassword(backend string) bool {
	for _, b := range passwordlessProviders {
		if b == backend {
			return false
		}
	}
	return true
}
