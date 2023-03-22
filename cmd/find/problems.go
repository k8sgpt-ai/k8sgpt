/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package find

import (
	"context"
	"os"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/client"
	"github.com/k8sgpt-ai/k8sgpt/pkg/openai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var explain bool

// problemsCmd represents the problems command
var problemsCmd = &cobra.Command{
	Use:   "problems",
	Short: "This command will find problems within your Kubernetes cluster",
	Long: `This command will find problems within your Kubernetes cluster and
	 provide you with a list of issues that need to be resolved`,
	Run: func(cmd *cobra.Command, args []string) {

		// Initialise the openAI client
		openAIClient, err := openai.NewClient()
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		// Get kubernetes client from viper
		client := viper.Get("kubernetesClient").(*client.Client)

		analyzer.RunAnalysis(ctx, client, openAIClient, explain)
	},
}

func init() {

	problemsCmd.Flags().BoolVarP(&explain, "explain", "e", false, "Explain the problem to me")

	FindCmd.AddCommand(problemsCmd)

}
