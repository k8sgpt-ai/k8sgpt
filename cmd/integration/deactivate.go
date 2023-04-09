/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package integration

import (
	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/integration"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// deactivateCmd represents the deactivate command
var deactivateCmd = &cobra.Command{
	Use:   "deactivate",
	Short: "Deactivate an integration",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		// Check if the integation exists
		integration := viper.Get("integration").(*integration.Integration)

		if err := integration.Deactivate(name, namespace); err != nil {
			color.Red("Error: %v", err)
			return
		}

		color.Green("Deactivate integration %s", name)

	},
}

func init() {
	IntegrationCmd.AddCommand(deactivateCmd)
}
