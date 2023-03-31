package filters

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available filters",
	Long:  `The list command displays a list of available filters that can be used to analyze Kubernetes resources.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Available filters : \n")
		for _, analyzer := range analyzer.ListFilters() {
			fmt.Printf("> %s\n", color.GreenString(analyzer))
		}
	},
}
