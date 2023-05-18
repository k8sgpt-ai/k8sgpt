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
	"github.com/spf13/viper"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a remote cache",
	Long:  `This command allows you to remove a remote cache and use the default filecache.`,
	Run: func(cmd *cobra.Command, args []string) {

		// Remove the remote cache
		var cacheInfo cache.CacheProvider
		err := viper.UnmarshalKey("cache", &cacheInfo)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}
		if cacheInfo.BucketName == "" {
			color.Yellow("Error: no cache is configured")
			os.Exit(1)
		}
		// Warn user this will delete the S3 bucket and prompt them to continue
		color.Yellow("Warning: this will delete the S3 bucket %s", cacheInfo.BucketName)
		color.Yellow("Are you sure you want to continue? (y/n)")
		var response string
		_, err = fmt.Scanln(&response)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}
		if response != "y" {
			os.Exit(0)
		}

	},
}

func init() {
	CacheCmd.AddCommand(removeCmd)
	removeCmd.Flags().StringVarP(&bucketname, "bucket", "b", "", "The name of the bucket to use for the cache")
}
