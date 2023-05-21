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

		for _, b := range inputBackends {
			foundBackend := false
			for i, provider := range configAI.Providers {
				if b == provider.Name {
					foundBackend = true
					configAI.Providers[i].Name = backend
					configAI.Providers[i].Model = model
					configAI.Providers[i].Password = password
					configAI.Providers[i].BaseURL = baseURL
					configAI.Providers[i].Engine = engine
					color.Green("%s updated in the AI backend provider list", b)
					break
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
	updateCmd.Flags().StringVarP(&backend, "backend", "b", "openai", "Backend AI provider")
	// update flag for model
	updateCmd.Flags().StringVarP(&model, "model", "m", "gpt-3.5-turbo", "Backend AI model")
	// update flag for password
	updateCmd.Flags().StringVarP(&password, "password", "p", "", "Backend AI password")
	// update flag for url
	updateCmd.Flags().StringVarP(&baseURL, "baseurl", "u", "", "URL AI provider, (e.g `http://localhost:8080/v1`)")
	// update flag for azure open ai engine/deployment name
	updateCmd.Flags().StringVarP(&engine, "engine", "e", "", "Azure AI deployment name")
}
