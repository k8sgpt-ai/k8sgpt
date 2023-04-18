package config

import (
	"fmt"
	"os"

	"github.com/mitchellh/mapstructure"
)

const (
	envPasswordType = "environment"
)

var _ PasswordProvider = (*EnvPasswordProvider)(nil)

func parseEnvPasswordProvider(data map[string]interface{}) (EnvPasswordProvider, error) {
	var result EnvPasswordProvider
	err := mapstructure.Decode(data, &result)

	if err != nil {
		return EnvPasswordProvider{}, err
	}

	return result, err
}

type EnvPasswordProvider struct {
	Name string `mapstructure:"name"`
}

func (s EnvPasswordProvider) GetPassword() (string, error) {
	value, found := os.LookupEnv(s.Name)

	if !found {
		return "", fmt.Errorf("cannot find env variable named `%s`", s.Name)
	}

	return value, nil
}
