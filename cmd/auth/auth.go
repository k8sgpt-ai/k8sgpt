package auth

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var (
	backend  string
	model    string
	password string
)

// authCmd represents the auth command
var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with your chosen backend",
	Long:  `Provide the necessary credentials to authenticate with your chosen backend.`,
	Run: func(cmd *cobra.Command, args []string) {

		configAI := ai.AIConfiguration{
			Providers: []ai.AIProvider{},
		}

		defaultProvider := ai.AIProvider{}

		backendType := viper.GetString("backend_type")
		if backendType == "" {
			// Set the default backend
			defaultProvider.Name = "openai"
		}
		// override the default backend if a flag is provided
		if backend != "" {
			defaultProvider.Name = backend
			color.Green("Using %s as backend AI provider", backendType)
		}

		if model != "" {
			defaultProvider.Model = model
		}

		if password == "" {
			fmt.Printf("Enter %s Key: ", backendType)
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				color.Red("Error reading %s Key from stdin: %s", backendType,
					err.Error())
				os.Exit(1)
			}
			password = strings.TrimSpace(string(bytePassword))
		}

		defaultProvider.Password = password

		configAI.Providers = append(configAI.Providers, defaultProvider)
		viper.Set("ai", configAI)
		if err := viper.WriteConfig(); err != nil {
			color.Red("Error writing config file: %s", err.Error())
			os.Exit(1)
		}
		color.Green("key added")
	},
}

func init() {
	// add flag for backend
	AuthCmd.Flags().StringVarP(&backend, "backend", "b", "openai", "Backend AI provider")
	// add flag for model
	AuthCmd.Flags().StringVarP(&model, "model", "m", "gpt-3.5-turbo", "Backend AI model")
	// add flag for password
	AuthCmd.Flags().StringVarP(&password, "password", "p", "", "Backend AI password")
}
