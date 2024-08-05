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
	"os"

	"github.com/fatih/color"
	customAnalyzer "github.com/k8sgpt-ai/k8sgpt/pkg/custom_analyzer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	name string
	url  string
	port int
)

var addCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"add"},
	Short:   "This command will add a custom analyzer from source",
	Long:    "This command allows you to add/remote/list an existing custom analyzer.",
	Run: func(cmd *cobra.Command, args []string) {
		err := viper.UnmarshalKey("custom_analyzers", &configCustomAnalyzer)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}
		analyzer := customAnalyzer.NewCustomAnalyzer()

		// Check if configuration is valid
		err = analyzer.Check(configCustomAnalyzer, name, url, port)
		if err != nil {
			color.Red("Error adding custom analyzer: %s", err.Error())
			os.Exit(1)
		}

		configCustomAnalyzer = append(configCustomAnalyzer, customAnalyzer.CustomAnalyzerConfiguration{
			Name: name,
			Connection: customAnalyzer.Connection{
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
	addCmd.Flags().StringVarP(&name, "name", "n", "my-custom-analyzer", "Name of the custom analyzer.")
	addCmd.Flags().StringVarP(&url, "url", "u", "localhost", "URL for the custom analyzer connection.")
	addCmd.Flags().IntVarP(&port, "port", "r", 8085, "Port for the custom analyzer connection.")
}
