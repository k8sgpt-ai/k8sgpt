package filters

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/spf13/cobra"
)

var FiltersCmd = &cobra.Command{
	Use:     "filters",
	Aliases: []string{"filters"},
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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available filters",
	Long:  `The list command displays a list of available filters that can be used to analyze Kubernetes resources.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Available filters : \n")
		for _, analyzer := range analyzer.ListAnalayzers() {
			fmt.Printf("> %s\n", color.GreenString(analyzer))
		}
	},
}

func init() {
	FiltersCmd.AddCommand(listCmd)
}
