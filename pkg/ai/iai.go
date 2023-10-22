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

	"github.com/k8sgpt-ai/k8sgpt/pkg/cache"
)

var (
	clients = []IAI{
		&OpenAIClient{},
		&AzureAIClient{},
		&LocalAIClient{},
		&NoOpAIClient{},
		&CohereClient{},
		&AmazonBedRockClient{},
	}
	Backends = []string{
		"openai",
		"localai",
		"azureopenai",
		"noopai",
		"cohere",
		"amazonbedrock",
	}
)

type IAI interface {
	Configure(config IAIConfig, language string) error
	GetCompletion(ctx context.Context, prompt string, promptTmpl string) (string, error)
	Parse(ctx context.Context, prompt []string, cache cache.ICache, promptTmpl string) (string, error)
	GetName() string
}

type IAIConfig interface {
	GetPassword() string
	GetModel() string
	GetBaseURL() string
	GetEngine() string
	GetTemperature() float32
	GetProviderRegion() string
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
	Name           string  `mapstructure:"name"`
	Model          string  `mapstructure:"model"`
	Password       string  `mapstructure:"password" yaml:"password,omitempty"`
	BaseURL        string  `mapstructure:"baseurl" yaml:"baseurl,omitempty"`
	Engine         string  `mapstructure:"engine" yaml:"engine,omitempty"`
	Temperature    float32 `mapstructure:"temperature" yaml:"temperature,omitempty"`
	ProviderRegion string  `mapstructure:"providerregion" yaml:"providerregion,omitempty"`
}

func (p *AIProvider) GetBaseURL() string {
	return p.BaseURL
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

func NeedPassword(backend string) bool {
	return backend != "localai"
}
