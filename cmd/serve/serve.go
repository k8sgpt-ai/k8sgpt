package serve

import (
	"fmt"
	"github.com/fatih/color"
	server2 "github.com/k8sgpt-ai/k8sgpt/pkg/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
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

		backendType := viper.GetString("backend_type")
		if backendType == "" {
			color.Red("No backend set. Please run k8sgpt auth")
			os.Exit(1)
		}

		if backend != "" {
			backendType = backend
		}

		token := viper.GetString(fmt.Sprintf("%s_key", backendType))
		// check if nil
		if token == "" {
			color.Red("No %s key set. Please run k8sgpt auth", backendType)
			os.Exit(1)
		}

		server := server2.K8sGPTServer{
			Backend: backend,
			Port:    port,
			Token:   token,
		}

		err := server.Serve()
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
