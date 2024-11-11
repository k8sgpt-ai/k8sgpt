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

package dump

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/version"
)

type K8sGPTInfo struct {
	Version string
	Commit  string
	Date    string
}
type DumpOut struct {
	AIConfiguration        ai.AIConfiguration
	ActiveFilters          []string
	KubenetesServerVersion *version.Info
	K8sGPTInfo             K8sGPTInfo
}

var DumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Creates a dumpfile for debugging issues with K8sGPT",
	Long:  `The dump command will create a dump.*.json which will contain K8sGPT non-sensitive configuration information.`,
	Run: func(cmd *cobra.Command, args []string) {

		// Fetch the configuration object(s)
		// get ai configuration
		var configAI ai.AIConfiguration
		err := viper.UnmarshalKey("ai", &configAI)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		var newProvider []ai.AIProvider
		for _, config := range configAI.Providers {
			// we blank out the custom headers for data protection reasons
			config.CustomHeaders = make([]http.Header, 0)
			// blank out the password
			if len(config.Password) > 4 {
				config.Password = config.Password[:4] + "***"
			} else {
				// If the password is shorter than 4 characters
				config.Password = "***"
			}
			newProvider = append(newProvider, config)
		}
		configAI.Providers = newProvider
		activeFilters := viper.GetStringSlice("active_filters")
		kubecontext := viper.GetString("kubecontext")
		kubeconfig := viper.GetString("kubeconfig")
		client, err := kubernetes.NewClient(kubecontext, kubeconfig)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}
		v, err := client.Client.Discovery().ServerVersion()
		if err != nil {
			color.Yellow("Could not find kubernetes server version")
		}
		var dumpOut DumpOut = DumpOut{
			AIConfiguration:        configAI,
			ActiveFilters:          activeFilters,
			KubenetesServerVersion: v,
			K8sGPTInfo: K8sGPTInfo{
				Version: viper.GetString("Version"),
				Commit:  viper.GetString("Commit"),
				Date:    viper.GetString("Date"),
			},
		}
		// Serialize dumpOut to JSON
		jsonData, err := json.MarshalIndent(dumpOut, "", " ")
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}
		// Write JSON data to file
		f := fmt.Sprintf("dump_%s.json", time.Now().Format("20060102150405"))
		err = os.WriteFile(f, jsonData, 0644)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}
		color.Green("Dump created successfully: %s", f)
	},
}

func init() {

}
