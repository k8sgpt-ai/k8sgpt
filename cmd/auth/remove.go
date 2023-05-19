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

var removeCmd = &cobra.Command{
	Use:   "remove [backend(s)]",
	Short: "Remove a provider",
	Long:  "The command to remove an AI backend provider",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputBackends := strings.Split(args[0], ",")
		err := viper.UnmarshalKey("ai", &configAI)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		if len(inputBackends) == 0 {
			color.Red("Error: backend must be set.")
			os.Exit(1)
		}

		for _, b := range inputBackends {
			foundBackend := false
			for i, provider := range configAI.Providers {
				if b == provider.Name {
					foundBackend = true
					configAI.Providers = append(configAI.Providers[:i], configAI.Providers[i+1:]...)
					color.Green("%s deleted to the AI backend provider list", b)
					break
				}
			}
			if !foundBackend {
				color.Red("Error: %s does not exist in configuration file. Please use k8sgpt auth new.", backend)
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

}

