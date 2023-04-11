package integration

import (
	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/integration"
	"github.com/spf13/cobra"
)

// activateCmd represents the activate command
var activateCmd = &cobra.Command{
	Use:   "activate [integration]",
	Short: "Activate an integration",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		integrationName := args[0]

		integration := integration.NewIntegration()
		// Check if the integation exists
		err := integration.Activate(integrationName, namespace)
		if err != nil {
			color.Red("Error: %v", err)
			return
		}

		color.Green("Activated integration %s", integrationName)
	},
}

func init() {
	IntegrationCmd.AddCommand(activateCmd)

}
