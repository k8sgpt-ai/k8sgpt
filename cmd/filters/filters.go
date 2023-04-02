package filters

import (
	"github.com/spf13/cobra"
)

var FiltersCmd = &cobra.Command{
	Use:     "filters",
	Aliases: []string{"filters", "filter"},
	Short:   "Manage filters for analyzing Kubernetes resources",
	Long: `The filters command allows you to manage filters that are used to analyze Kubernetes resources.
	You can list available filters to analyze resources.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
	},
}

func init() {
	FiltersCmd.AddCommand(listCmd)
	FiltersCmd.AddCommand(addCmd)
	FiltersCmd.AddCommand(removeCmd)
}
