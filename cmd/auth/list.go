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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured providers",
	Long:  "The list command displays a list of configured providers",
	Run: func(cmd *cobra.Command, args []string) {

		// get ai configuration
		err := viper.UnmarshalKey("ai", &configAI)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		// iterate over the provider list and prints each provider name
		for _, provider := range configAI.Providers {
			if len(configAI.Providers) == 0 {
				color.Red("Provider list is currently empty.")
			} else {
				fmt.Printf("> %s\n", provider.Name)
			}
		}
	},
}
