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
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/version"
)

type DumpOut struct {
	AIConfiguration        ai.AIConfiguration
	ActiveFilters          []string
	KubenetesServerVersion *version.Info
}

var DumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Creates a dump for debugging",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		// Fetch the configuration object(s)
		// get ai configuration
		var configAI ai.AIConfiguration
		err := viper.UnmarshalKey("ai", &configAI)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		for _, config := range configAI.Providers {
			// blank out the password
			config.Password = ""
		}

		activeFilters := viper.GetStringSlice("active_filters")
		// Get Kubernetes server data
		kubecontext := viper.GetString("kubecontext")
		kubeconfig := viper.GetString("kubeconfig")
		client, err := kubernetes.NewClient(kubecontext, kubeconfig)

		version, err := client.Client.Discovery().ServerVersion()
		if err != nil {
			color.Yellow("Could not find kubernetes server version")
		}
		var dumpOut DumpOut = DumpOut{
			AIConfiguration:        configAI,
			ActiveFilters:          activeFilters,
			KubenetesServerVersion: version,
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
