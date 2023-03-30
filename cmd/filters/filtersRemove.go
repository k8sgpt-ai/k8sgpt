package filters

import (
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var filtersRemoveCmd = &cobra.Command{
	Use:   "remove [filter(s)]",
	Short: "Remove one or more filters.",
	Long:  `The add command remove one or more filters to the default set of filters used by the analyze.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		// Get defined default_filters
		defaultFilters := viper.GetStringSlice("default_filters")
		if len(defaultFilters) == 0 {
			defaultFilters = []string{}
		}

		// verify dupplicate filters example: k8sgpt filters remove Pod Pod
		uniqueFilters, dupplicateFilters := util.RemoveDuplicates(args)
		if len(dupplicateFilters) != 0 {
			color.Red("Duplicate filters found: %s", strings.Join(dupplicateFilters, ", "))
			os.Exit(1)
		}

		// Verify if filter exist in config file and update default_filter
		filterNotFound := []string{}
		for _, filter := range uniqueFilters {
			foundFilter := false
			for i, f := range defaultFilters {
				if f == filter {
					foundFilter = true
					defaultFilters = append(defaultFilters[:i], defaultFilters[i+1:]...)
					break
				}
			}
			if !foundFilter {
				filterNotFound = append(filterNotFound, filter)
			}
		}

		if len(filterNotFound) != 0 {
			color.Red("Filter(s) %s does not exist in configuration file. Please use k8sgpt filters add.", strings.Join(filterNotFound, ", "))
			os.Exit(1)
		}

		viper.Set("default_filters", defaultFilters)

		if err := viper.WriteConfig(); err != nil {
			color.Red("Error writing config file: %s", err.Error())
			os.Exit(1)
		}
		color.Green("Filter(s) %s removed", strings.Join(args, ", "))
	},
}
