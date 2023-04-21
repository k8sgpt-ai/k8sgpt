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
	"github.com/k8sgpt-ai/k8sgpt/pkg/integration"
	"github.com/spf13/cobra"
)

// deactivateCmd represents the deactivate command
var deactivateCmd = &cobra.Command{
	Use:   "deactivate [integration]",
	Short: "Deactivate an integration",
	Args:  cobra.ExactArgs(1),
	Long:  `For example e.g. k8sgpt integration deactivate trivy`,
	Run: func(cmd *cobra.Command, args []string) {
		integrationName := args[0]

		integration := integration.NewIntegration()

		if err := integration.Deactivate(integrationName, namespace); err != nil {
			color.Red("Error: %v", err)
			return
		}

		color.Green("Deactivated integration %s", integrationName)

	},
}

func init() {
	IntegrationCmd.AddCommand(deactivateCmd)
}
