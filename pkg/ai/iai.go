package ai

import (
	"context"
)

type IAIConfig interface {
	GetToken() string
	GetModel() string
	GetBaseURL() string
	GetEngine() string
}

type IAI interface {
	Configure(config IAIConfig, language string) error
	GetCompletion(ctx context.Context, prompt string) (string, error)
	Parse(ctx context.Context, prompt []string, nocache bool) (string, error)
	GetName() string
}

func NewClient(provider string) IAI {
	switch provider {
	case "openai":
		return &OpenAIClient{}
	case "noopai":
		return &NoOpAIClient{}
	case "azureopenai":
		return &AzureAIClient{}
	case "llama":
		return &LLaMAAIClient{}
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
	Password string `mapstructure:"password" yaml:"password,omitempty"`
	BaseURL  string `mapstructure:"baseurl" yaml:"baseurl,omitempty"`
	Engine   string `mapstructure:"engine" yaml:"engine,omitempty"`
}

func (p *AIProvider) GetToken() string {
	return p.Password
}

func (p *AIProvider) GetModel() string {
	return p.Model
}

func (p *AIProvider) GetBaseURL() string {
	return p.BaseURL
}

func (p *AIProvider) GetEngine() string {
	return p.Engine
}

func NeedPassword(backend string) bool {
	if backend == "llama" {
		return false
	}
	return true
}
