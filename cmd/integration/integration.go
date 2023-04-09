package integration

import (
	"github.com/spf13/cobra"
)

var (
	namespace string
	name      string
)

// IntegrationCmd represents the integrate command
var IntegrationCmd = &cobra.Command{
	Use:   "integration",
	Short: "Intergrate another tool into K8sGPT",
	Long: `Intergrate another tool into K8sGPT. For example:
	k8sgpt integration add trivy
	
	This would allow you to deploy trivy into your cluster and use a K8sGPT analyzer to parse trivy results.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	IntegrationCmd.PersistentFlags().StringVarP(&name, "name", "a", "", "The name of the integration")
	IntegrationCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "The namespace to use for the integration")
}
