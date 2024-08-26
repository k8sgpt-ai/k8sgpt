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
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	names string
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove custom analyzer(s)",
	Long:  "The command to remove custom analyzer(s)",
	PreRun: func(cmd *cobra.Command, args []string) {
		// Ensure that the "names" flag is provided before running the command
		_ = cmd.MarkFlagRequired("names")
	},
	Run: func(cmd *cobra.Command, args []string) {
		if names == "" {
			// Display an error message and show command help if "names" is not set
			color.Red("Error: names must be set.")
			_ = cmd.Help()
			return
		}
		// Split the provided names by comma
		inputCustomAnalyzers := strings.Split(names, ",")

		// Load the custom analyzers from the configuration file
		err := viper.UnmarshalKey("custom_analyzers", &configCustomAnalyzer)
		if err != nil {
			// Display an error message if the configuration cannot be loaded
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		// Iterate over each input analyzer name
		for _, inputAnalyzer := range inputCustomAnalyzers {
			foundAnalyzer := false
			// Search for the analyzer in the current configuration
			for i, analyzer := range configCustomAnalyzer {
				if analyzer.Name == inputAnalyzer {
					foundAnalyzer = true

					// Remove the analyzer from the configuration list
					configCustomAnalyzer = append(configCustomAnalyzer[:i], configCustomAnalyzer[i+1:]...)
					color.Green("%s deleted from the custom analyzer list", analyzer.Name)
					break
				}
			}
			if !foundAnalyzer {
				// Display an error if the analyzer is not found in the configuration
				color.Red("Error: %s does not exist in configuration file. Please use k8sgpt custom-analyzer add.", inputAnalyzer)
				os.Exit(1)
			}
		}

		// Save the updated configuration back to the file
		viper.Set("custom_analyzers", configCustomAnalyzer)
		if err := viper.WriteConfig(); err != nil {
			// Display an error if the configuration cannot be written
			color.Red("Error writing config file: %s", err.Error())
			os.Exit(1)
		}

	},
}

func init() {
	// add flag for names
	removeCmd.Flags().StringVarP(&names, "names", "n", "", "Custom analyzers to remove (separated by a comma)")
}
