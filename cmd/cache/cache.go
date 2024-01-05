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
	"github.com/spf13/cobra"
)

var (
	bucketname string
)

// cacheCmd represents the cache command
var CacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "For working with the cache the results of an analysis",
	Long:  `Cache commands allow you to add a remote cache, list the contents of the cache, and remove items from the cache.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			panic(err)
		}
	},
}

func init() {
}
