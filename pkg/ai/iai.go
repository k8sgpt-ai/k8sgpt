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

// clientConstructors maps each backend name to a constructor that returns a
// fresh, zero-value client instance on every call. This replaces the previous
// package-level slice of shared pointers, which caused a data race when
// concurrent gRPC Query requests called Configure() on the same instance.
var clientConstructors = map[string]func() IAI{
	openAIClientName:                func() IAI { return &OpenAIClient{} },
	anthropicClientName:             func() IAI { return &AnthropicClient{} },
	azureAIClientName:               func() IAI { return &AzureAIClient{} },
	localAIClientName:               func() IAI { return &LocalAIClient{} },
	ollamaClientName:                func() IAI { return &OllamaClient{} },
	noopAIClientName:                func() IAI { return &NoOpAIClient{} },
	cohereAIClientName:              func() IAI { return &CohereClient{} },
	amazonbedrockAIClientName:       func() IAI { return &AmazonBedRockClient{} },
	amazonBedrockConverseClientName: func() IAI { return &AmazonBedrockConverseClient{} },
	amazonsagemakerAIClientName:     func() IAI { return &SageMakerAIClient{} },
	googleAIClientName:              func() IAI { return &GoogleGenAIClient{} },
	huggingfaceAIClientName:         func() IAI { return &HuggingfaceClient{} },
	googleVertexAIClientName:        func() IAI { return &GoogleVertexAIClient{} },
	ociClientName:                   func() IAI { return &OCIGenAIClient{} },
	CustomRestClientName:            func() IAI { return &CustomRestClient{} },
	ibmWatsonxAIClientName:          func() IAI { return &IBMWatsonxAIClient{} },
	groqAIClientName:                func() IAI { return &GroqClient{} },
}

var Backends = []string{
	openAIClientName,
	anthropicClientName,
	localAIClientName,
	ollamaClientName,
	azureAIClientName,
	cohereAIClientName,
	amazonbedrockAIClientName,
	amazonBedrockConverseClientName,
	amazonsagemakerAIClientName,
	googleAIClientName,
	noopAIClientName,
	huggingfaceAIClientName,
	googleVertexAIClientName,
	ociClientName,
	CustomRestClientName,
	ibmWatsonxAIClientName,
	groqAIClientName,
}

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
	GetStopSequences() []string
	GetProviderId() string
	GetCompartmentId() string
	GetOrganizationId() string
	GetAzureAPIType() string
	GetAzureAPIVersion() string
	GetCustomHeaders() []http.Header
}

// NewClient returns a fresh IAI instance for the given provider name on every
// invocation. Callers (including concurrent gRPC handlers) may safely call
// Configure() on the returned value without any shared-state data race.
func NewClient(provider string) IAI {
	if constructor, ok := clientConstructors[provider]; ok {
		return constructor()
	}
	// default client
	return &OpenAIClient{}
}

type AIConfiguration struct {
	Providers       []AIProvider `mapstructure:"providers"`
	DefaultProvider string       `mapstructure:"defaultprovider"`
}

type AIProvider struct {
	Name            string        `mapstructure:"name"`
	Model           string        `mapstructure:"model"`
	Password        string        `mapstructure:"password" yaml:"password,omitempty"`
	BaseURL         string        `mapstructure:"baseurl" yaml:"baseurl,omitempty"`
	ProxyEndpoint   string        `mapstructure:"proxyEndpoint" yaml:"proxyEndpoint,omitempty"`
	ProxyPort       string        `mapstructure:"proxyPort" yaml:"proxyPort,omitempty"`
	EndpointName    string        `mapstructure:"endpointname" yaml:"endpointname,omitempty"`
	Engine          string        `mapstructure:"engine" yaml:"engine,omitempty"`
	Temperature     float32       `mapstructure:"temperature" yaml:"temperature,omitempty"`
	ProviderRegion  string        `mapstructure:"providerregion" yaml:"providerregion,omitempty"`
	ProviderId      string        `mapstructure:"providerid" yaml:"providerid,omitempty"`
	CompartmentId   string        `mapstructure:"compartmentid" yaml:"compartmentid,omitempty"`
	TopP            float32       `mapstructure:"topp" yaml:"topp,omitempty"`
	TopK            int32         `mapstructure:"topk" yaml:"topk,omitempty"`
	MaxTokens       int           `mapstructure:"maxtokens" yaml:"maxtokens,omitempty"`
	StopSequences   []string      `mapstructure:"stopsequences" yaml:"stopsequences,omitempty"`
	OrganizationId  string        `mapstructure:"organizationid" yaml:"organizationid,omitempty"`
	AzureAPIType    string        `mapstructure:"azureapitype" yaml:"azureapitype,omitempty"`
	AzureAPIVersion string        `mapstructure:"azureapiversion" yaml:"azureapiversion,omitempty"`
	CustomHeaders   []http.Header `mapstructure:"customHeaders"`
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

func (p *AIProvider) GetStopSequences() []string {
	return p.StopSequences
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

func (p *AIProvider) GetAzureAPIType() string {
	return p.AzureAPIType
}

func (p *AIProvider) GetAzureAPIVersion() string {
	return p.AzureAPIVersion
}

func (p *AIProvider) GetCustomHeaders() []http.Header {
	return p.CustomHeaders
}

var passwordlessProviders = []string{"localai", "ollama", "amazonsagemaker", "amazonbedrock", "amazonbedrockconverse", "googlevertexai", "oci", "customrest"}

func NeedPassword(backend string) bool {
	for _, b := range passwordlessProviders {
		if b == backend {
			return false
		}
	}
	return true
}
