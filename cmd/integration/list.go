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
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/integration"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists built-in integrations",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		integrationProvider := integration.NewIntegration()
		integrations := integrationProvider.List()

		fmt.Println(color.YellowString("Active:"))
		for _, i := range integrations {
			b, err := integrationProvider.IsActivate(i)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if b {
				fmt.Printf("> %s\n", color.GreenString(i))
			}
		}

		fmt.Println(color.YellowString("Unused: "))
		for _, i := range integrations {
			b, err := integrationProvider.IsActivate(i)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if !b {
				fmt.Printf("> %s\n", color.GreenString(i))
			}
		}
	},
}

func init() {
	IntegrationCmd.AddCommand(listCmd)

}
