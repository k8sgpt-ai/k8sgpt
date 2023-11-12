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
package cache

import (
	"os"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/cache"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove the remote cache",
	Long:  `This command allows you to remove the remote cache and use the default filecache.`,
	Run: func(cmd *cobra.Command, args []string) {

		err := cache.RemoveRemoteCache()
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}
		color.Green("Successfully removed the remote cache")
	},
}

func init() {
	CacheCmd.AddCommand(removeCmd)
}
