package auth

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/lockandkey"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var (
	backend    string
	password   string
	model      string
	passphrase string
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

		var encryptedPassword string
		if password == "" {
			fmt.Printf("Enter %s Key: ", backend)
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				color.Red("Error reading %s Key from stdin: %s", backend,
					err.Error())
				os.Exit(1)
			}
			password = strings.TrimSpace(string(bytePassword))
		}

		var key string
		if passphrase == "" {
			fmt.Printf("\nEnter Passphrase for the API Key: ")
			bytePassphrase, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				color.Red("Error reading Passphrase for the API Key Key from stdin: %s",
					err.Error())
				os.Exit(1)
			}
			passphrase = strings.TrimSpace(string(bytePassphrase))
			if passphrase != "" {
				key = passphrase
				if len(key) != 16 {
					color.Red("Encryption passphrase is not of suitable lenght of 16")
					os.Exit(1)
				}
				encryptionKey := []byte(key)
				//encrypting password
				encryptedPassword, err = lockandkey.Encrypt(encryptionKey, []byte(password))
				if err != nil {
					color.Red("Encryption of API key failed with: %s",
						err.Error())
					os.Exit(1)
				}
			} else {
				key = ""
				encryptedPassword = password
			}
		}

		// create new provider object
		newProvider := ai.AIProvider{
			Name:       backend,
			Model:      model,
			Password:   encryptedPassword,
			Passphrase: key,
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
	// add flag for passphrase
	AuthCmd.Flags().StringVarP(&passphrase, "with-passphrase", "e", "", "Passphrase(of lenght 16) for encryption of API Key")
}
