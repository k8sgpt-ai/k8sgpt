package config

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type persistedAIConfiguration struct {
	Providers []persistedAIProvider `mapstructure:"providers"`
}

type persistedAIProvider struct {
	Name     string      `mapstructure:"name"`
	Model    string      `mapstructure:"model"`
	Password interface{} `mapstructure:"password"`
	BaseURL  string      `mapstructure:"base_url"`
}

func LoadAIConfiguration() (AIConfiguration, error) {
	var configAI persistedAIConfiguration
	err := viper.UnmarshalKey("ai", &configAI)

	if err != nil {
		return AIConfiguration{}, err
	}

	return toAIConfiguration(configAI)
}

func toAIConfiguration(persisted persistedAIConfiguration) (AIConfiguration, error) {
	var aiProviders []AIProvider

	for _, provider := range persisted.Providers {
		result, err := toAIProvider(provider)

		if err != nil {
			return AIConfiguration{}, err
		}

		aiProviders = append(aiProviders, result)
	}

	return AIConfiguration{
		Providers: aiProviders,
	}, nil
}

func toAIProvider(persisted persistedAIProvider) (AIProvider, error) {
	passwordProvider, err := toPasswordProvider(persisted.Password)

	if err != nil {
		return AIProvider{}, err
	}

	return AIProvider{
		Name:     persisted.Name,
		Model:    persisted.Model,
		Password: passwordProvider,
		BaseURL:  persisted.BaseURL,
	}, nil
}

type passwordProviderSpec struct {
	Type      string                 `mapstructure:"type"`
	Remaining map[string]interface{} `mapstructure:",remain"`
}

func toPasswordProvider(password interface{}) (PasswordProvider, error) {
	switch pwd := password.(type) {
	case map[string]interface{}:
		var spec passwordProviderSpec

		err := mapstructure.Decode(pwd, &spec)

		if err != nil {
			return nil, err
		}

		return toPasswordProviderForType(spec.Type, spec.Remaining)
	case string:
		return SimpleTextPasswordProvider{Password: pwd}, nil
	}

	return nil, fmt.Errorf("cannot get password provider from: %s", password)
}

func toPasswordProviderForType(providerType string, data map[string]interface{}) (PasswordProvider, error) {
	switch providerType {
	case envPasswordType:
		return parseEnvPasswordProvider(data)
	case commandPasswordType:
		return parseCommandPasswordProvider(data)
	}

	return nil, fmt.Errorf("unknown provider type `%s`", providerType)
}

func SetAIConfig(aiConfig AIConfiguration) {
	persisted := toPersistedAIConfiguration(aiConfig)

	viper.Set("ai", persisted)
}

func toPersistedAIConfiguration(aiConfig AIConfiguration) persistedAIConfiguration {
	var providers []persistedAIProvider

	for _, p := range aiConfig.Providers {
		persistedProvider := toPersistedAIProvider(p)

		providers = append(providers, persistedProvider)
	}

	return persistedAIConfiguration{
		Providers: providers,
	}
}

func toPersistedAIProvider(aiProvider AIProvider) persistedAIProvider {
	password := toPersistedPassword(aiProvider.Password)

	return persistedAIProvider{
		Name:     aiProvider.Name,
		Model:    aiProvider.Model,
		Password: password,
	}
}

func toPersistedPassword(passwordProvider PasswordProvider) interface{} {
	switch p := passwordProvider.(type) {
	case SimpleTextPasswordProvider:
		return p.Password
	case EnvPasswordProvider:
		return map[string]string{
			"type": envPasswordType,
			"name": p.Name,
		}
	case CommandPasswordProvider:
		return map[string]interface{}{
			"type":      envPasswordType,
			"command":   p.Command,
			"arguments": p.Arguments,
		}
	}

	fmt.Println("warning: unknown password provider! cannot persist!")
	return nil
}
