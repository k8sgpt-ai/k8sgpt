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

package filters

import (
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var addCmd = &cobra.Command{
	Use:   "add [filter(s)]",
	Short: "Adds one or more new filters.",
	Long:  `The add command adds one or more new filters to the default set of filters used by the analyze.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputFilters := strings.Split(args[0], ",")
		coreFilters, additionalFilters, integrationFilters := analyzer.ListFilters()

		availableFilters := append(append(coreFilters, additionalFilters...), integrationFilters...)
		// Verify filter exist
		invalidFilters := []string{}
		for _, f := range inputFilters {
			if f == "" {
				color.Red("Filter cannot be empty. Please use correct syntax.")
				os.Exit(1)
			}
			foundFilter := false
			for _, filter := range availableFilters {
				if filter == f {
					foundFilter = true

					// WARNING: This is to enable users correctly understand implications
					// of enabling logs
					if filter == "Log" {
						color.Yellow("Warning: by enabling logs, you will be sending potentially sensitive data to the AI backend.")
					}

					break
				}
			}
			if !foundFilter {
				invalidFilters = append(invalidFilters, f)
			}
		}

		if len(invalidFilters) != 0 {
			color.Red("Filter %s does not exist. Please use k8sgpt filters list", strings.Join(invalidFilters, ", "))
			os.Exit(1)
		}

		// Get defined active_filters
		activeFilters := viper.GetStringSlice("active_filters")
		if len(activeFilters) == 0 {
			activeFilters = coreFilters
		}

		mergedFilters := append(activeFilters, inputFilters...)

		uniqueFilters, dupplicatedFilters := util.RemoveDuplicates(mergedFilters)

		// Verify dupplicate
		if len(dupplicatedFilters) != 0 {
			color.Red("Duplicate filters found: %s", strings.Join(dupplicatedFilters, ", "))
			os.Exit(1)
		}

		viper.Set("active_filters", uniqueFilters)

		if err := viper.WriteConfig(); err != nil {
			color.Red("Error writing config file: %s", err.Error())
			os.Exit(1)
		}
		color.Green("Filter %s added", strings.Join(inputFilters, ", "))
	},
}
