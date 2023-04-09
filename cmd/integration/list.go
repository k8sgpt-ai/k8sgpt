package integration

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/integration"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists built-in integrations",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		integration := viper.Get("integration").(*integration.Integration)
		integrations := integration.List()

		for _, integration := range integrations {
			fmt.Printf("> %s\n", color.GreenString(integration))
		}
	},
}

func init() {
	IntegrationCmd.AddCommand(listCmd)

}
