package serve

import (
	"os"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	k8sgptserver "github.com/k8sgpt-ai/k8sgpt/pkg/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	port    string
	backend string
	token   string
)

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Runs k8sgpt as a server",
	Long:  `Runs k8sgpt as a server to allow for easy integration with other applications.`,
	Run: func(cmd *cobra.Command, args []string) {

		var configAI ai.AIConfiguration
		err := viper.UnmarshalKey("ai", &configAI)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		if len(configAI.Providers) == 0 {
			color.Red("Error: AI provider not specified in configuration. Please run k8sgpt auth")
			os.Exit(1)
		}

		var aiProvider ai.AIProvider
		for _, provider := range configAI.Providers {
			if backend == provider.Name {
				aiProvider = provider
				break
			}
		}

		if aiProvider.Name == "" {
			color.Red("Error: AI provider %s not specified in configuration. Please run k8sgpt auth", backend)
			os.Exit(1)
		}

		server := k8sgptserver.Config{
			Backend: aiProvider.Name,
			Port:    port,
			Token:   aiProvider.Password,
		}

		err = server.Serve()
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}
		// override the default backend if a flag is provided
	},
}

func init() {
	// add flag for backend
	ServeCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to run the server on")
	ServeCmd.Flags().StringVarP(&backend, "backend", "b", "openai", "Backend AI provider")
}
