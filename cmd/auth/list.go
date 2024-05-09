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

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var details bool

func runListCommand(cmd *cobra.Command, args []string) {
	// Get the ai configurations
	err := viper.UnmarshalKey("ai", &configAI)
	if err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}

	// Print the default if it is set
	fmt.Print(color.YellowString("Default: \n"))
	if configAI.DefaultProvider != "" {
		fmt.Printf("> %s\n", color.BlueString(configAI.DefaultProvider))
	} else {
		fmt.Printf("> %s\n", color.BlueString("openai"))
	}

	// Get list of all AI Backends and only print them if they are not in the provider list
	fmt.Print(color.YellowString("Active: \n"))
	for _, provider := range configAI.Providers {
		fmt.Printf("> %s\n", color.GreenString(provider.Backend))
		fmt.Println("  > " + color.HiCyanString("Configs"))

		for index, config := range provider.Configs {
			if index == provider.DefaultConfig {
				fmt.Printf("    %d. %s "+color.HiYellowString("(Default Config)\n"), index+1, config.Name)
			} else {
				fmt.Printf("    %d. %s\n", index+1, config.Name)
			}
			if details {
				printDetails(provider, index)
			}
		}
	}

	fmt.Print(color.YellowString("Unused: \n"))
	for _, aiBackend := range ai.Backends {
		providerExists := false
		for _, provider := range configAI.Providers {
			if provider.Backend == aiBackend {
				providerExists = true
			}
		}
		if !providerExists {
			fmt.Printf("> %s\n", color.RedString(aiBackend))
		}
	}
}

func printDetails(provider ai.AIProvider, index int) {
	if provider.Configs[index].Model != "" {
		fmt.Printf("       - Model: %s\n", provider.Configs[index].Model)
	}

	if provider.Configs[index].Engine != "" {
		fmt.Printf("       - Engine: %s\n", provider.Configs[index].Engine)
	}

	if provider.Configs[index].BaseURL != "" {
		fmt.Printf("       - BaseURL: %s\n", provider.Configs[index].BaseURL)
	}
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured providers",
	Long:  "The list command displays a list of configured providers",
	Run:   runListCommand,
}

func init() {
	listCmd.Flags().BoolVar(&details, "details", false, "Print active provider configuration details")
}
