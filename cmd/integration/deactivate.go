/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package integration

import (
	"github.com/k8sgpt-ai/k8sgpt/pkg/integration"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	name string
)

// deactivateCmd represents the deactivate command
var deactivateCmd = &cobra.Command{
	Use:   "deactivate",
	Short: "Deactivate an integration",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		// Check if the integation exists
		integration := viper.Get("integration").(*integration.Integration)

		integration.Deactivate(name, namespace)

		// Deactivate
	},
}

func init() {
	IntegrationCmd.AddCommand(deactivateCmd)
	deactivateCmd.Flags().StringVarP(&name, "name", "n", "", "The name of the integration to deactivate")
}
