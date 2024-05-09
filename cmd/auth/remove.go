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

func runRemoveCommand(cmd *cobra.Command, args []string) {
	// Get the ai configurations
	err := viper.UnmarshalKey("ai", &configAI)
	if err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}

	// Check if the backend flag is set.
	if backend == "" {
		color.Red("Error: backends must be set.")
		_ = cmd.Help()
		return
	}

	inputBackends := strings.Split(backend, ",")

	if configName == "" {
		color.Yellow("Warning: No config is specified therefore the default config will be removed")
	}

	// Now, iterate over each backend
	for _, backendName := range inputBackends {
		foundBackend := false
		for i, provider := range configAI.Providers {
			// Check if the input backend is present in the list of providers stored
			// in the config file.
			if backendName == provider.Backend {
				foundBackend = true

				// Now, start iterating over the configs stored in the backend
				deletedConfigIndex := -1

				if configName == "" {
					// Delete the current default config if no config name has been specified.
					deletedConfigIndex = provider.DefaultConfig
					configName = provider.Configs[provider.DefaultConfig].Name
				} else {
					for index, config := range provider.Configs {
						if configName == config.Name {
							deletedConfigIndex = index
							break
						}
					}
				}

				if deletedConfigIndex != -1 {
					// Remove the config if it is found.
					configAI.Providers[i].Configs = append(configAI.Providers[i].Configs[:deletedConfigIndex], configAI.Providers[i].Configs[deletedConfigIndex+1:]...)
					color.Green("Config: \"%s\" deleted for the AI backend provider: \"%s\"", configName, backendName)
				} else {
					color.Red("Error: Backend provider \"%s\" didn't have any config with name \"%s\". Aborting!", backendName, configName)
					os.Exit(1)
				}

				// Now, check if there are any configs left for this backend provider.
				if len(configAI.Providers[i].Configs) == 0 {
					// Delete this backend provider.
					configAI.Providers = append(configAI.Providers[:i], configAI.Providers[i+1:]...)

					// Check if this was also the default provider.
					if configAI.DefaultProvider == backendName {
						// Update the default backend to the first item in the backend list
						if len(configAI.Providers) != 0 {
							configAI.DefaultProvider = configAI.Providers[0].Backend
						} else {
							// If there are no providers left. Fallback to default.
							configAI.DefaultProvider = defaultBackend
						}
					}
				} else if deletedConfigIndex == configAI.Providers[i].DefaultConfig {
					// Update the default config for this backend provider to config at 0th index.
					configAI.Providers[i].DefaultConfig = 0
					color.Yellow("After deleting the config \"%s\", the default config for backend provider \"%s\" has changed to \"%s\"", configName, backendName, configAI.Providers[i].Configs[0].Name)
				}

				break
			}
		}
		if !foundBackend {
			color.Red("Error: \"%s\" does not exist in configuration file. Please use k8sgpt auth new.", backendName)
			os.Exit(1)
		}
	}

	viper.Set("ai", configAI)
	if err := viper.WriteConfig(); err != nil {
		color.Red("Error writing config file: %s", err.Error())
		os.Exit(1)
	}
}

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove provider(s)",
	Long:  "The command to remove AI backend provider(s)",
	PreRun: func(cmd *cobra.Command, args []string) {
		_ = cmd.MarkFlagRequired("backends")
	},
	Run: runRemoveCommand,
}

func init() {
	// add flag for backends
	removeCmd.Flags().StringVarP(&backend, "backends", "b", "", "Backend AI providers to remove (separated by a comma)")
	removeCmd.Flags().StringVarP(&configName, "config-name", "", "", "Name of the config to remove")
}
