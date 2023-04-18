package config

import (
	"io"
	"os/exec"

	"github.com/mitchellh/mapstructure"
)

const (
	commandPasswordType = "command"
)

var _ PasswordProvider = (*CommandPasswordProvider)(nil)

type CommandPasswordProvider struct {
	Command   string   `mapstructure:"command"`
	Arguments []string `mapstructure:"arguments"`
}

func (r CommandPasswordProvider) GetPassword() (string, error) {
	cmd := exec.Command(r.Command, r.Arguments...)
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	data, err := io.ReadAll(stdout)

	if err != nil {
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		return "", err
	}

	return string(data), nil
}

func parseCommandPasswordProvider(data map[string]interface{}) (CommandPasswordProvider, error) {
	var result CommandPasswordProvider
	err := mapstructure.Decode(data, &result)

	if err != nil {
		return CommandPasswordProvider{}, err
	}

	return result, err
}
