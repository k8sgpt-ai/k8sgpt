package config

import (
	"github.com/spf13/viper"
)

func LoadAIConfiguration() (AIConfiguration, error) {
	var configAI AIConfiguration
	err := viper.UnmarshalKey("ai", &configAI)

	if err != nil {
		return AIConfiguration{}, err
	}

	return configAI, nil
}

func SetAIConfig(aiConfig AIConfiguration) {
	viper.Set("ai", aiConfig)
}
