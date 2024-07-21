/*
Copyright 2023 The K8sGPT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package customanalyzer

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	customanalyzer "github.com/k8sgpt-ai/k8sgpt/pkg/customAnalyzer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	installType string
	install     bool
	packageUrl  string
	name        string
	url         string
	port        int
)

var AddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"add"},
	Short:   "This command will add a custom analyzer from source",
	Long:    "This command allows you to add a custom analyzer from a specified source and optionally install it.",
	PreRun: func(cmd *cobra.Command, args []string) {
		if install {
			_ = cmd.MarkFlagRequired("install-type")
			_ = cmd.MarkFlagRequired("package")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := viper.UnmarshalKey("custom_analyzers", &configCustomAnalyzer)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}
		customAnalyzer := customanalyzer.NewCustomAnalyzer()

		// Check if configuration is valid
		err = customAnalyzer.Check(configCustomAnalyzer, name, url, port)
		if err != nil {
			color.Red("Error adding custom analyzer: %s", err.Error())
			os.Exit(1)
		}

		if install {
			// Check if installType is possible
			install, err := customAnalyzer.GetInstallType(installType)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// create a pod in-cluster with custom analyzer
			err = install.Deploy(packageUrl, name, url, port)
			if err != nil {
				color.Red("Error installing custom analyzer: %s", err.Error())
				os.Exit(1)
			}
		}

		configCustomAnalyzer = append(configCustomAnalyzer, customanalyzer.CustomAnalyzerConfiguration{
			Name: name,
			Connection: customanalyzer.Connection{
				Url:  url,
				Port: port,
			},
		})

		viper.Set("custom_analyzers", configCustomAnalyzer)
		if err := viper.WriteConfig(); err != nil {
			color.Red("Error writing config file: %s", err.Error())
			os.Exit(1)
		}
		color.Green("%s added to the custom analyzers config list", name)

	},
}

func init() {
	AddCmd.Flags().StringVarP(&installType, "install-type", "t", "docker", "Specify the installation type (e.g., docker, kubernetes).")
	AddCmd.Flags().BoolVarP(&install, "install", "i", false, "Flag to indicate whether to install the custom analyzer after adding.")
	AddCmd.Flags().StringVarP(&packageUrl, "package", "p", "", "URL of the custom analyzer package.")
	AddCmd.Flags().StringVarP(&name, "name", "n", "my-custom-analyzer", "Name of the custom analyzer.")
	AddCmd.Flags().StringVarP(&url, "url", "u", "localhost", "URL for the custom analyzer connection.")
	AddCmd.Flags().IntVarP(&port, "port", "r", 8085, "Port for the custom analyzer connection.")
}
