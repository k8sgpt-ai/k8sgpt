package config

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/fatih/color"
	"github.com/spf13/viper"
)

type IAIConfig interface {
	GetPassword() (string, error)
	GetModel() string
	GetBaseURL() string
}

type AIConfiguration struct {
	Providers []AIProvider
}

type PasswordProvider interface {
	GetPassword() (string, error)
}

type AIProvider struct {
	Name     string
	Model    string
	Password PasswordProvider
	BaseURL  string
}

func (p *AIProvider) GetBaseURL() string {
	return p.BaseURL
}

func (p *AIProvider) GetPassword() (string, error) {
	return p.Password.GetPassword()
}

func (p *AIProvider) GetModel() string {
	return p.Model
}

func GetBackendType() string {
	return viper.GetString("backend_type")
}

func PersistOrFail() {
	if err := viper.WriteConfig(); err != nil {
		color.Red("Error writing config file: %s", err.Error())
		os.Exit(1)
	}
}

func Initialize(cfgFile string) {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// the config will belocated under `~/.config/k8sgpt/k8sgpt.yaml` on linux
		configDir := filepath.Join(xdg.ConfigHome, "k8sgpt")

		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("k8sgpt")

		// nothing we can really do in case of an error
		_ = viper.SafeWriteConfig()
	}

	viper.SetEnvPrefix("K8SGPT")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in. Nothing we can really do in case of an error
	_ = viper.ReadInConfig()
}
