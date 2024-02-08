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
	newConfigName string
)

func runUpdateCommand(cmd *cobra.Command, args []string) {
	// Get the ai configurations
	err := viper.UnmarshalKey("ai", &configAI)
	if err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}

	// Check if the backend flag is set.
	if backend == "" {
		color.Red("Error: backend must be set.")
		_ = cmd.Help()
		return
	}

	// Validate the temperature range.
	if temperature > 1.0 || temperature < 0.0 {
		color.Red("Error: temperature ranges from 0 to 1.")
		os.Exit(1)
	}

	// Iterate over all the providers present in the config file.
	for i, provider := range configAI.Providers {
		if backend == provider.Backend {
			configIndex := -1

			if configName == "" {
				// Modify the default config if the config name is not specified.
				color.Yellow("Since no config name was specified, changes will be made to the default config")
				configIndex = provider.DefaultConfig
				configName = provider.Configs[provider.DefaultConfig].Name
			} else {
				// Iterate over all the configs present in that backend provider.
				for index, config := range provider.Configs {
					// Check if the config to be updated exists or not.
					if configName == config.Name {
						configIndex = index
					}
				}
			}

			if configIndex == -1 {
				color.Red("Error: The backend provider \"%s\" does not have a configuration with the name \"%s\"", backend, configName)
				os.Exit(1)
			} else {
				// Config exists, now update the parameters
				if newConfigName != "" {
					configAI.Providers[i].Configs[configIndex].Name = newConfigName
					color.Blue("Config name updated successfully")
				}
				if model != "" {
					configAI.Providers[i].Configs[configIndex].Model = model
					color.Blue("Model updated successfully")
				}
				if password != "" {
					configAI.Providers[i].Configs[configIndex].Password = password
					color.Blue("Password updated successfully")
				}
				if baseURL != "" {
					configAI.Providers[i].Configs[configIndex].BaseURL = baseURL
					color.Blue("Base URL updated successfully")
				}
				if engine != "" {
					configAI.Providers[i].Configs[configIndex].Engine = engine
					color.Blue("Engine updated successfully")
				}
				configAI.Providers[i].Configs[configIndex].Temperature = temperature
				color.Green("Config \"%s\" for the backend provider \"%s\" has been successfully updated", configName, backend)
			}

			// Break out of the loop if the desired backend provider has been updated.
			break
		}
	}

	// Write the configuration to the config file.
	viper.Set("ai", configAI)
	if err := viper.WriteConfig(); err != nil {
		color.Red("Error writing config file: %s", err.Error())
		os.Exit(1)
	}
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a backend provider",
	Long:  "The command to update an AI backend provider",
	// TODO: Why was this present in the first place?
	// Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		backend, _ := cmd.Flags().GetString("backend")
		if strings.ToLower(backend) == "azureopenai" {
			_ = cmd.MarkFlagRequired("engine")
			_ = cmd.MarkFlagRequired("baseurl")
		}
	},
	Run: runUpdateCommand,
}

func init() {
	// update flag for backend
	updateCmd.Flags().StringVarP(&backend, "backend", "b", "", "Update backend AI provider")
	// update flag for config-name
	updateCmd.Flags().StringVarP(&configName, "config-name", "", "", "Name of the configuration to update")
	// update flag for config-name
	updateCmd.Flags().StringVarP(&newConfigName, "name", "n", "", "New name for the configuration to update")
	// update flag for model
	updateCmd.Flags().StringVarP(&model, "model", "m", "", "Update backend AI model")
	// update flag for password
	updateCmd.Flags().StringVarP(&password, "password", "p", "", "Update backend AI password")
	// update flag for url
	updateCmd.Flags().StringVarP(&baseURL, "baseurl", "u", "", "Update URL AI provider, (e.g `http://localhost:8080/v1`)")
	// add flag for temperature
	updateCmd.Flags().Float32VarP(&temperature, "temperature", "t", 0.7, "The sampling temperature, value ranges between 0 ( output be more deterministic) and 1 (more random)")
	// update flag for azure open ai engine/deployment name
	updateCmd.Flags().StringVarP(&engine, "engine", "e", "", "Update Azure AI deployment name")
}
