/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package find

import (
	"github.com/spf13/cobra"
)

// problemsCmd represents the problems command
var problemsCmd = &cobra.Command{
	Use:   "problems",
	Short: "This command will find problems within your Kubernetes cluster",
	Long: `This command will find problems within your Kubernetes cluster and
	 provide you with a list of issues that need to be resolved`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	FindCmd.AddCommand(problemsCmd)

}
