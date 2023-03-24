package find

import (
	"github.com/spf13/cobra"
)

// findCmd represents the find command
var FindCmd = &cobra.Command{
	Use:   "find",
	Short: "Find issues within your Kubernetes cluster",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {

}
