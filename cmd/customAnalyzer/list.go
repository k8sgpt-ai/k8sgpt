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
	"fmt"
	"os"

	"github.com/fatih/color"
	customAnalyzer "github.com/k8sgpt-ai/k8sgpt/pkg/custom_analyzer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var details bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured custom analyzers",
	Long:  "The list command displays a list of configured custom analyzers",
	Run: func(cmd *cobra.Command, args []string) {

		// get custom_analyzers configuration
		err := viper.UnmarshalKey("custom_analyzers", &configCustomAnalyzer)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		// Get list of all Custom Analyers configured
		fmt.Print(color.YellowString("Active: \n"))
		for _, analyzer := range configCustomAnalyzer {
			fmt.Printf("> %s\n", color.GreenString(analyzer.Name))
			if details {
				printDetails(analyzer)
			}
		}
	},
}

func init() {
	listCmd.Flags().BoolVar(&details, "details", false, "Print custom analyzers configuration details")
}

func printDetails(analyzer customAnalyzer.CustomAnalyzerConfiguration) {
	fmt.Printf("   - Url: %s\n", analyzer.Connection.Url)
	fmt.Printf("   - Port: %d\n", analyzer.Connection.Port)

}
