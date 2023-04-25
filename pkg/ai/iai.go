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

type IAI interface {
	Configure(config IAIConfig, language string) error
	GetCompletion(ctx context.Context, prompt string) (string, error)
	Parse(ctx context.Context, prompt []string, cache cache.ICache) (string, error)
	GetName() string
}

type IAIConfig interface {
	GetPassword() string
	GetModel() string
	GetBaseURL() string
}

func NewClient(provider string) IAI {
	switch provider {
	case "openai":
		return &OpenAIClient{}
	case "localai":
		return &LocalAIClient{}
	case "noopai":
		return &NoOpAIClient{}
	default:
		return &OpenAIClient{}
	}
}

type AIConfiguration struct {
	Providers []AIProvider `mapstructure:"providers"`
}

type AIProvider struct {
	Name     string `mapstructure:"name"`
	Model    string `mapstructure:"model"`
	Password string `mapstructure:"password"`
	BaseURL  string `mapstructure:"baseurl"`
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
