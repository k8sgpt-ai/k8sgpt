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

var defaultCmd = &cobra.Command{
	Use:   "default",
	Short: "Set your default AI backend provider",
	Long:  "The command to set your new default AI backend provider (default is openai)",
	Run: func(cmd *cobra.Command, args []string) {
		err := viper.UnmarshalKey("ai", &configAI)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}
		if providerName == "" {
			if configAI.DefaultProvider != "" {
				color.Yellow("Your default provider is %s", configAI.DefaultProvider)
			} else {
				color.Yellow("Your default provider is openai")
			}
			os.Exit(0)
		}
		// lowercase the provider name
		providerName = strings.ToLower(providerName)

		// Check if the provider is in the provider list
		providerExists := false
		for _, provider := range configAI.Providers {
			if provider.Name == providerName {
				providerExists = true
			}
		}
		if !providerExists {
			color.Red("Error: Provider %s does not exist", providerName)
			os.Exit(1)
		}
		// Set the default provider
		configAI.DefaultProvider = providerName

		viper.Set("ai", configAI)
		// Viper write config
		err = viper.WriteConfig()
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}
		// Print acknowledgement
		color.Green("Default provider set to %s", providerName)
	},
}

func init() {
	// provider name flag
	defaultCmd.Flags().StringVarP(&providerName, "provider", "p", "", "The name of the provider to set as default")
}
