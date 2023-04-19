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
	password string
	model    string
	baseURL  string
	azEngine string
)

// authCmd represents the auth command
var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with your chosen backend",
	Long:  `Provide the necessary credentials to authenticate with your chosen backend.`,
	Run: func(cmd *cobra.Command, args []string) {

		// get ai configuration
		var configAI ai.AIConfiguration
		err := viper.UnmarshalKey("ai", &configAI)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		// search for provider with same name
		providerIndex := -1
		for i, provider := range configAI.Providers {
			if backend == provider.Name {
				providerIndex = i
				break
			}
		}

		// check if backend is not empty
		if backend == "" {
			color.Red("Error: Backend AI cannot be empty.")
			os.Exit(1)
		}

		color.Green("Using %s as backend AI provider", backend)

		// check if model is not empty
		if model == "" {
			color.Red("Error: Model cannot be empty.")
			os.Exit(1)
		}

		if ai.NeedPassword(backend) && password == "" {
			fmt.Printf("Enter %s Key: ", backend)
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				color.Red("Error reading %s Key from stdin: %s", backend,
					err.Error())
				os.Exit(1)
			}
			password = strings.TrimSpace(string(bytePassword))
		}

		// create new provider object
		newProvider := ai.AIProvider{
			Name:     backend,
			Model:    model,
			Password: password,
			BaseURL:  baseURL,
			Engine:   azEngine,
		}

		if providerIndex == -1 {
			// provider with same name does not exist, add new provider to list
			configAI.Providers = append(configAI.Providers, newProvider)
			color.Green("New provider added")
		} else {
			// provider with same name exists, update provider info
			configAI.Providers[providerIndex] = newProvider
			color.Green("Provider updated")
		}
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
	// add flag for url
	AuthCmd.Flags().StringVarP(&baseURL, "baseurl", "u", "", "URL AI provider, (e.g `http://localhost:8080/v1`)")
	// add flag for azure open ai engine/deployment name
	AuthCmd.Flags().StringVarP(&azEngine, "engine", "e", "", "Azure AI deployment name")
}
