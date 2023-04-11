package filters

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/integration"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available filters",
	Long:  `The list command displays a list of available filters that can be used to analyze Kubernetes resources.`,
	Run: func(cmd *cobra.Command, args []string) {
		activeFilters := viper.GetStringSlice("active_filters")
		coreFilters, additionalFilters, integrationFilters := analyzer.ListFilters()

		availableFilters := append(coreFilters, additionalFilters...)
		availableFilters = append(availableFilters, integrationFilters...)
		if len(activeFilters) == 0 {
			activeFilters = coreFilters
		}
		inactiveFilters := util.SliceDiff(availableFilters, activeFilters)
		fmt.Printf(color.YellowString("Active: \n"))
		for _, filter := range activeFilters {
			fmt.Printf("> %s\n", color.GreenString(filter))
		}

		// Add integrations ( which are dynamic ) to active filters
		integrationProvider := integration.NewIntegration()
		fmt.Printf(color.BlueString("Active Integrations: \n"))
		for _, filter := range integrationFilters {
			b, err := integrationProvider.IsActivate(filter)
			if err != nil {
				fmt.Printf(color.RedString("Error: %s", err))
			}
			if b {
				fmt.Printf("> %s\n", color.GreenString(filter))
			}
		}

		// display inactive filters
		if len(inactiveFilters) != 0 {
			fmt.Printf(color.YellowString("Unused: \n"))
			for _, filter := range inactiveFilters {
				fmt.Printf("> %s\n", color.RedString(filter))
			}
		}

	},
}
