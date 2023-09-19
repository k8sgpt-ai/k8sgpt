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

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a backend provider",
	Long:  "The command to update an AI backend provider",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		backend, _ := cmd.Flags().GetString("backend")
		if strings.ToLower(backend) == "azureopenai" {
			_ = cmd.MarkFlagRequired("engine")
			_ = cmd.MarkFlagRequired("baseurl")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {

		// get ai configuration
		err := viper.UnmarshalKey("ai", &configAI)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		inputBackends := strings.Split(args[0], ",")

		if len(inputBackends) == 0 {
			color.Red("Error: backend must be set.")
			os.Exit(1)
		}
		if temperature > 1.0 || temperature < 0.0 {
			color.Red("Error: temperature ranges from 0 to 1.")
			os.Exit(1)
		}

		for _, b := range inputBackends {
			foundBackend := false
			for i, provider := range configAI.Providers {
				if b == provider.Name {
					foundBackend = true
					if backend != "" {
						configAI.Providers[i].Name = backend
						color.Blue("Backend name updated successfully")
					}
					if model != "" {
						configAI.Providers[i].Model = model
						color.Blue("Model updated successfully")
					}
					if password != "" {
						configAI.Providers[i].Password = password
						color.Blue("Password updated successfully")
					}
					if baseURL != "" {
						configAI.Providers[i].BaseURL = baseURL
						color.Blue("Base URL updated successfully")
					}
					if engine != "" {
						configAI.Providers[i].Engine = engine
					}
					configAI.Providers[i].Temperature = temperature
					color.Green("%s updated in the AI backend provider list", b)
				}
			}
			if !foundBackend {
				color.Red("Error: %s does not exist in configuration file. Please use k8sgpt auth new.", args[0])
				os.Exit(1)
			}

		}

		viper.Set("ai", configAI)
		if err := viper.WriteConfig(); err != nil {
			color.Red("Error writing config file: %s", err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	// update flag for backend
	updateCmd.Flags().StringVarP(&backend, "backend", "b", "", "Update backend AI provider")
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
