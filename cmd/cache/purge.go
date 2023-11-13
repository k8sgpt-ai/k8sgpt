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
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/cache"
	"github.com/spf13/cobra"
)

var purgeCmd = &cobra.Command{
	Use:   "purge [object name]",
	Short: "Purge a remote cache",
	Long:  "This command allows you to delete/purge one object from the cache",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			color.Red("Error: Please provide a value for object name. Run k8sgpt cache purge --help")
			os.Exit(1)
		}
		objectKey := args[0]
		fmt.Println(color.YellowString("Purging a remote cache."))
		c, err := cache.GetCacheConfiguration()
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		err = c.Remove(objectKey)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}
		fmt.Println(color.GreenString("Object deleted."))
	},
}

func init() {
	CacheCmd.AddCommand(purgeCmd)
}
