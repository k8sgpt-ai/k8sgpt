package cmd

import (
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of k8sgpt",
	Long:  `All software has versions. This is k8sgpt's`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("k8sgpt version %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
