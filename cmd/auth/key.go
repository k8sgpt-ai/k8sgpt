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

// keyCmd represents the key command
var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Add a key to OpenAI",
	Long:  `This command will add a key from OpenAI to enable you to interact with the API`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Print("Enter OpenAI API Key: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			color.Red("Error reading OpenAI API Key from stdin: %s", err.Error())
			os.Exit(1)
		}
		password := strings.TrimSpace(string(bytePassword))

		viper.Set("openai_api_key", password)
		if err := viper.WriteConfig(); err != nil {
			color.Red("Error writing config file: %s", err.Error())
			os.Exit(1)
		}
		color.Green("key added")
	},
}

func init() {
	AuthCmd.AddCommand(keyCmd)

}
