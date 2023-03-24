package auth

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

// authCmd represents the auth command
var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with your chosen backend",
	Long:  `Provide the necessary credentials to authenticate with your chosen backend.`,
	Run: func(cmd *cobra.Command, args []string) {

		backendType := viper.GetString("backend_type")

		fmt.Printf("Enter %s Key: ", backendType)
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			color.Red("Error reading %s Key from stdin: %s", backendType,
				err.Error())
			os.Exit(1)
		}
		password := strings.TrimSpace(string(bytePassword))

		viper.Set(fmt.Sprintf("%s_key", backendType), password)
		if err := viper.WriteConfig(); err != nil {
			color.Red("Error writing config file: %s", err.Error())
			os.Exit(1)
		}
		color.Green("key added")
	},
}

func init() {

}
