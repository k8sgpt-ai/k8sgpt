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
		filters := strings.Split(args[0], ",")

		// Verify filter exist
		invalidFilters := []string{}
		for _, f := range filters {
			if f == "" {
				color.Red("Filter cannot be empty. Please use correct syntax.")
				os.Exit(1)
			}
			foundFilter := false
			for _, filter := range analyzer.ListFilters() {
				if filter == f {
					foundFilter = true
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

		// Get defined default_filters
		defaultFilters := viper.GetStringSlice("default_filters")
		if len(defaultFilters) == 0 {
			defaultFilters = []string{}
		}

		mergedFilters := append(defaultFilters, filters...)

		uniqueFilters, dupplicatedFilters := util.RemoveDuplicates(mergedFilters)

		// Verify dupplicate
		if len(dupplicatedFilters) != 0 {
			color.Red("Duplicate filters found: %s", strings.Join(dupplicatedFilters, ", "))
			os.Exit(1)
		}

		viper.Set("default_filters", uniqueFilters)

		if err := viper.WriteConfig(); err != nil {
			color.Red("Error writing config file: %s", err.Error())
			os.Exit(1)
		}
		color.Green("Filter %s added", strings.Join(filters, ", "))
	},
}
