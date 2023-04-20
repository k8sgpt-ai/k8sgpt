package ai

import (
	"context"
)

type IAI interface {
	Configure(config IAIConfig, language string) error
	GetCompletion(ctx context.Context, prompt string) (string, error)
	Parse(ctx context.Context, prompt []string, nocache bool) (string, error)
	GetName() string
}

type IAIConfig interface {
	GetPassword() string
	GetModel() string
	GetBaseURL() string
	GetEngine() string
}

func NewClient(provider string) IAI {
	switch provider {
	case "openai":
		return &OpenAIClient{}
	case "azureopenai":
		return &AzureAIClient{}
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
	Password string `mapstructure:"password" yaml:"password,omitempty"`
	BaseURL  string `mapstructure:"baseurl" yaml:"baseurl,omitempty"`
	Engine   string `mapstructure:"engine" yaml:"engine,omitempty"`
}

func (p *AIProvider) GetPassword() string {
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
