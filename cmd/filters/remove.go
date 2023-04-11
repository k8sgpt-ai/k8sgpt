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

var removeCmd = &cobra.Command{
	Use:   "remove [filter(s)]",
	Short: "Remove one or more filters.",
	Long:  `The add command remove one or more filters to the default set of filters used by the analyze.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputFilters := strings.Split(args[0], ",")

		// Get defined active_filters
		activeFilters := viper.GetStringSlice("active_filters")
		coreFilters, _, _ := analyzer.ListFilters()

		if len(activeFilters) == 0 {
			activeFilters = coreFilters
		}

		// Check if input input filters is not empty
		for _, f := range inputFilters {
			if f == "" {
				color.Red("Filter cannot be empty. Please use correct syntax.")
				os.Exit(1)
			}
		}

		// verify dupplicate filters example: k8sgpt filters remove Pod Pod
		uniqueFilters, dupplicatedFilters := util.RemoveDuplicates(inputFilters)
		if len(dupplicatedFilters) != 0 {
			color.Red("Duplicate filters found: %s", strings.Join(dupplicatedFilters, ", "))
			os.Exit(1)
		}

		// Verify if filter exist in config file and update default_filter
		filterNotFound := []string{}
		for _, filter := range uniqueFilters {
			foundFilter := false
			for i, f := range activeFilters {
				if f == filter {
					foundFilter = true
					activeFilters = append(activeFilters[:i], activeFilters[i+1:]...)
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

		viper.Set("active_filters", activeFilters)

		if err := viper.WriteConfig(); err != nil {
			color.Red("Error writing config file: %s", err.Error())
			os.Exit(1)
		}
		color.Green("Filter(s) %s removed", strings.Join(inputFilters, ", "))
	},
}
