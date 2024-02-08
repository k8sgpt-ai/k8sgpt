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

package auth

import (
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	providerName string
)

func runDefaultCommand(cmd *cobra.Command, args []string) {
	// 1. Get the ai configurations
	err := viper.UnmarshalKey("ai", &configAI)
	if err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}

	// 2. Validate the input values and set defaults if necessary
	if providerName == "" {
		if configAI.DefaultProvider != "" {
			color.Yellow("Your default provider is \"%s\"", configAI.DefaultProvider)
		} else {
			color.Yellow("Your default provider is openai")
		}
		os.Exit(0)
	}

	// lowercase the provider name
	providerName = strings.ToLower(providerName)

	// Check if the provider is in the provider list
	providerIndex := -1
	configIndex := -1
	for i, provider := range configAI.Providers {
		if providerName == provider.Backend {
			providerIndex = i

			// Iterate over all the configs of this provider
			// and check if a config with the same name exists
			if configName != "" {
				for index, config := range provider.Configs {
					if configName == config.Name {
						configIndex = index
						break
					}
				}
			}

			if configIndex != -1 {
				break
			}
		}
	}

	if providerIndex == -1 {
		color.Red("Error: Provider \"%s\" does not exist", providerName)
		os.Exit(1)
	}

	if configIndex == -1 && configName != "" {
		color.Red("Error: The backend provider \"%s\" does not have a configuration with the name \"%s\"", backend, configName)
		os.Exit(1)
	}

	if configName != "" {
		// Set the default config
		configAI.Providers[providerIndex].DefaultConfig = configIndex
	} else {
		// Set the default provider
		configAI.DefaultProvider = providerName
	}

	viper.Set("ai", configAI)
	// Viper write config
	err = viper.WriteConfig()
	if err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}

	// Print acknowledgement
	if configName != "" {
		color.Green("Default config for %s set to %s", providerName, configName)
	} else {
		color.Green("Default provider set to %s", providerName)
	}
}

var defaultCmd = &cobra.Command{
	Use:   "default",
	Short: "Set your default AI backend provider and provider config",
	Long:  "The command to set your new default AI backend provider (default is openai)",
	Run:   runDefaultCommand,
}

func init() {
	// provider name flag
	defaultCmd.Flags().StringVarP(&providerName, "provider", "p", "", "The name of the provider to set as default")
	defaultCmd.Flags().StringVarP(&configName, "config-name", "", "", "The name of the config to set as default for a provider")
}
