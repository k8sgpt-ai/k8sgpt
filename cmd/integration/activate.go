package integration

import (
	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/integration"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// activateCmd represents the activate command
var activateCmd = &cobra.Command{
	Use:   "activate [integration]",
	Short: "Activate an integration",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		intName := args[0]

		integration := viper.Get("integration").(*integration.Integration)
		// Check if the integation exists
		err := integration.Activate(intName, namespace)
		if err != nil {
			color.Red("Error: %v", err)
			return
		}

		color.Yellow("Activating analyzer for integration %s", intName)

		// TODO:

		color.Green("Activate integration %s", intName)
	},
}

func init() {
	IntegrationCmd.AddCommand(activateCmd)

}
