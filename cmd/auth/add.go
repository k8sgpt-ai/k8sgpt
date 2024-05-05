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
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

const (
	defaultBackend = "openai"
	defaultModel   = "gpt-3.5-turbo"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add new provider",
	Long:  "The add command allows to configure a new backend AI provider",
	PreRun: func(cmd *cobra.Command, args []string) {
		backend, _ := cmd.Flags().GetString("backend")
		if strings.ToLower(backend) == "azureopenai" {
			_ = cmd.MarkFlagRequired("engine")
			_ = cmd.MarkFlagRequired("baseurl")
		}
		if strings.ToLower(backend) == "amazonsagemaker" {
			_ = cmd.MarkFlagRequired("endpointname")
			_ = cmd.MarkFlagRequired("providerRegion")
		}
		if strings.ToLower(backend) == "amazonbedrock" {
			_ = cmd.MarkFlagRequired("providerRegion")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {

		// get ai configuration
		err := viper.UnmarshalKey("ai", &configAI)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		// search for provider with same name
		providerIndex := -1
		for i, provider := range configAI.Providers {
			if backend == provider.Name {
				providerIndex = i
				break
			}
		}

		validBackend := func(validBackends []string, backend string) bool {
			for _, b := range validBackends {
				if b == backend {
					return true
				}
			}
			return false
		}

		// check if backend is not empty and a valid value
		if backend == "" {
			color.Yellow(fmt.Sprintf("Warning: backend input is empty, will use the default value: %s", defaultBackend))
			backend = defaultBackend
		} else {
			if !validBackend(ai.Backends, backend) {
				color.Red("Error: Backend AI accepted values are '%v'", strings.Join(ai.Backends, ", "))
				os.Exit(1)
			}
		}

		// check if model is not empty
		if model == "" {
			model = defaultModel
			color.Yellow(fmt.Sprintf("Warning: model input is empty, will use the default value: %s", defaultModel))
		}
		if temperature > 1.0 || temperature < 0.0 {
			color.Red("Error: temperature ranges from 0 to 1.")
			os.Exit(1)
		}
		if topP > 1.0 || topP < 0.0 {
			color.Red("Error: topP ranges from 0 to 1.")
			os.Exit(1)
		}

		if ai.NeedPassword(backend) && password == "" {
			fmt.Printf("Enter %s Key: ", backend)
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				color.Red("Error reading %s Key from stdin: %s", backend,
					err.Error())
				os.Exit(1)
			}
			password = strings.TrimSpace(string(bytePassword))
		}

		// create new provider object
		newProvider := ai.AIProvider{
			Name:           backend,
			Model:          model,
			Password:       password,
			BaseURL:        baseURL,
			EndpointName:   endpointName,
			Engine:         engine,
			Temperature:    temperature,
			ProviderRegion: providerRegion,
			ProviderId:     providerId,
			TopP:           topP,
			MaxTokens:      maxTokens,
		}

		if providerIndex == -1 {
			// provider with same name does not exist, add new provider to list
			configAI.Providers = append(configAI.Providers, newProvider)
			viper.Set("ai", configAI)
			if err := viper.WriteConfig(); err != nil {
				color.Red("Error writing config file: %s", err.Error())
				os.Exit(1)
			}
			color.Green("%s added to the AI backend provider list", backend)
		} else {
			// provider with same name exists, update provider info
			color.Yellow("Provider with same name already exists.")
		}
	},
}

func init() {
	// add flag for backend
	addCmd.Flags().StringVarP(&backend, "backend", "b", defaultBackend, "Backend AI provider")
	// add flag for model
	addCmd.Flags().StringVarP(&model, "model", "m", defaultModel, "Backend AI model")
	// add flag for password
	addCmd.Flags().StringVarP(&password, "password", "p", "", "Backend AI password")
	// add flag for url
	addCmd.Flags().StringVarP(&baseURL, "baseurl", "u", "", "URL AI provider, (e.g `http://localhost:8080/v1`)")
	// add flag for endpointName
	addCmd.Flags().StringVarP(&endpointName, "endpointname", "n", "", "Endpoint Name, e.g. `endpoint-xxxxxxxxxxxx` (only for amazonbedrock, amazonsagemaker backends)")
	// add flag for topP
	addCmd.Flags().Float32VarP(&topP, "topp", "c", 0.5, "Probability Cutoff: Set a threshold (0.0-1.0) to limit word choices. Higher values add randomness, lower values increase predictability.")
	// max tokens
	addCmd.Flags().IntVarP(&maxTokens, "maxtokens", "l", 2048, "Specify a maximum output length. Adjust (1-...) to control text length. Higher values produce longer output, lower values limit length")
	// add flag for temperature
	addCmd.Flags().Float32VarP(&temperature, "temperature", "t", 0.7, "The sampling temperature, value ranges between 0 ( output be more deterministic) and 1 (more random)")
	// add flag for azure open ai engine/deployment name
	addCmd.Flags().StringVarP(&engine, "engine", "e", "", "Azure AI deployment name (only for azureopenai backend)")
	//add flag for amazonbedrock region name
	addCmd.Flags().StringVarP(&providerRegion, "providerRegion", "r", "", "Provider Region name (only for amazonbedrock, googlevertexai backend)")
	//add flag for vertexAI Project ID
	addCmd.Flags().StringVarP(&providerId, "providerId", "i", "", "Provider specific ID for e.g. project (only for googlevertexai backend)")
}
