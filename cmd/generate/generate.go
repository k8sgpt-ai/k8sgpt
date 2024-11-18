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

package generate

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	backend     string
)

var apiKeysURLs = map[string]string{
	"openai":          "https://beta.openai.com/account/api-keys",
	"cohere":          "https://dashboard.cohere.ai/api-keys",
	"azureopenai":     "https://portal.azure.com/#create/Microsoft.CognitiveServicesOpenAI",
	"google":          "https://makersuite.google.com/app/apikey",
	"amazonsagemaker": "https://console.aws.amazon.com/bedrock/home",
	"amazonbedrock":   "https://console.aws.amazon.com/sagemaker/home",
	"huggingface":     "https://huggingface.co/settings/tokens",
	"localai":         "https://localai.io/basics/getting_started/",
	"noopai":          "https://docs.k8sgpt.ai/reference/providers/backend/#FakeAI",
}

// generateCmd represents the auth command
var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Key for your chosen backend (opens browser)",
	Long:  `Opens your browser to generate a key for your chosen backend.`,
	Run: func(cmd *cobra.Command, args []string) {
		if backend == "" {
			backend = "openai"
		}
		openbrowser(apiKeysURLs[backend])
	},
}

func init() {
	// add flag for backend
	GenerateCmd.Flags().StringVarP(&backend, "backend", "b", "openai", "Backend AI provider")
}

func openbrowser(url string) {
	var err error
	isGui := true
	switch runtime.GOOS {
	case "linux":
		_, err = exec.LookPath("xdg-open")
		if err != nil {
			isGui = false
		} else {
			err = exec.Command("xdg-open", url).Start()
		}
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	printInstructions(isGui, backend, url)
	if err != nil {
		fmt.Println(err)
	}
}

func printInstructions(isGui bool, backendType string, url string) {
	fmt.Println("")
	if isGui {
		color.Green("Opening: %s to generate a key for %s", url, backendType)
		fmt.Println("")
	} else {
		color.Green("Please open: %s to generate a key for %s", url, backendType)
		fmt.Println("")
	}
	color.Green("Please copy the generated key and run `k8sgpt auth add` to add it to your config file")
	fmt.Println("")
}
