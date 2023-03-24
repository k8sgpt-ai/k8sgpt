package init

import (
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	backend string
)

// authCmd represents the auth command
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise k8sgpt with a backend AI provider",
	Long:  `Currently only OpenAI is supported.`,
	Run: func(cmd *cobra.Command, args []string) {

		/*
			This is largely a placeholder for now. In the future we will support
			multiple backends and this will allow us to set the backend we want to use.
		*/
		if backend != "openai" {
			color.Yellow("Only OpenAI is supported at the moment")
		}
		viper.Set("backend_type", backend)
		if err := viper.WriteConfig(); err != nil {
			color.Red("Error writing config file: %s", err.Error())
			os.Exit(1)
		}

		color.Green("Backend set to %s", backend)
	},
}

func init() {
	// add flag for backend
	InitCmd.Flags().StringVarP(&backend, "backend", "b", "openai", "Backend AI provider")
}
