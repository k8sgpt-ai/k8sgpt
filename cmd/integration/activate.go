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

package integration

import (
	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/integration"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// activateCmd represents the activate command
var activateCmd = &cobra.Command{
	Use:   "activate [integration]",
	Short: "Activate an integration",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		integrationName := args[0]
		coreFilters, _, _ := analyzer.ListFilters()

		// Update filters
		activeFilters := viper.GetStringSlice("active_filters")
		if len(activeFilters) == 0 {
			activeFilters = coreFilters
		}

		integration := integration.NewIntegration()
		// Check if the integation exists
		err := integration.Activate(integrationName, namespace, activeFilters)
		if err != nil {
			color.Red("Error: %v", err)
			return
		}

		color.Green("Activated integration %s", integrationName)
	},
}

func init() {
	IntegrationCmd.AddCommand(activateCmd)

}
