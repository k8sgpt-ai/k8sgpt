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

package customanalyzer

import (
	customAnalyzer "github.com/k8sgpt-ai/k8sgpt/pkg/custom_analyzer"
	"github.com/spf13/cobra"
)

var configCustomAnalyzer []customAnalyzer.CustomAnalyzerConfiguration

// authCmd represents the auth command
var CustomAnalyzerCmd = &cobra.Command{
	Use:   "custom-analyzer",
	Short: "Manage a custom analyzer",
	Long:  `This command allows you to manage custom analyzers, including adding, removing, and listing them.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}
	},
}

func init() {
	// add subcommand to add custom analyzer
	CustomAnalyzerCmd.AddCommand(addCmd)
	// remove subcomment to remove custom analyzer
	CustomAnalyzerCmd.AddCommand(removeCmd)
	// list subcomment to list custom analyzer
	CustomAnalyzerCmd.AddCommand(listCmd)
}
