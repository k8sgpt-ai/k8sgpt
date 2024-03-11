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

package auth

import (
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/spf13/cobra"
)

var (
	backend        string
	password       string
	baseURL        string
	endpointName   string
	model          string
	engine         string
	temperature    float32
	providerRegion string
	providerId     string
	topP           float32
	maxTokens      int
)

var configAI ai.AIConfiguration

// authCmd represents the auth command
var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with your chosen backend",
	Long:  `Provide the necessary credentials to authenticate with your chosen backend.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}
	},
}

func init() {
	// add subcommand to list backends
	AuthCmd.AddCommand(listCmd)
	// add subcommand to create new backend provider
	AuthCmd.AddCommand(addCmd)
	// add subcommand to remove new backend provider
	AuthCmd.AddCommand(removeCmd)
	// add subcommand to set default backend provider
	AuthCmd.AddCommand(defaultCmd)
	// add subcommand to update backend provider
	AuthCmd.AddCommand(updateCmd)
}
