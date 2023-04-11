package integration

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/integration"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists built-in integrations",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		integrationProvider := integration.NewIntegration()
		integrations := integrationProvider.List()

		fmt.Println(color.YellowString("Active:"))
		for _, i := range integrations {
			b, err := integrationProvider.IsActivate(i)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if b {
				fmt.Printf("> %s\n", color.GreenString(i))
			}
		}

		fmt.Println(color.YellowString("Unused: "))
		for _, i := range integrations {
			b, err := integrationProvider.IsActivate(i)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if !b {
				fmt.Printf("> %s\n", color.GreenString(i))
			}
		}
	},
}

func init() {
	IntegrationCmd.AddCommand(listCmd)

}
