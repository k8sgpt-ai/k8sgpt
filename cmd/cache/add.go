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
	"strings"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/cache"
	"github.com/spf13/cobra"
)

var (
	region string
	//nolint:unused
	bucketName     string
	storageAccount string
	containerName  string
	projectId      string
	endpoint       string
	insecure       bool
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [cache type]",
	Short: "Add a remote cache",
	Long: `This command allows you to add a remote cache to store the results of an analysis.
	The supported cache types are:
	- Azure Blob storage (e.g., k8sgpt cache add azure)
	- Google Cloud storage (e.g., k8sgpt cache add gcs)
	- S3 (e.g., k8sgpt cache add s3)
	- Interplex (e.g., k8sgpt cache add interplex)`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			color.Red("Error: Please provide a value for cache types. Run k8sgpt cache add --help")
			os.Exit(1)
		}
		fmt.Println(color.YellowString("Adding remote based cache"))
		cacheType := args[0]
		remoteCache, err := cache.NewCacheProvider(strings.ToLower(cacheType), bucketName, region, endpoint, storageAccount, containerName, projectId, insecure)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}
		err = cache.AddRemoteCache(remoteCache)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	CacheCmd.AddCommand(addCmd)
	addCmd.Flags().StringVarP(&region, "region", "r", "us-east-1", "The region to use for the AWS S3 or GCS cache")
	addCmd.Flags().StringVarP(&endpoint, "endpoint", "e", "", "The S3 or minio endpoint")
	addCmd.Flags().BoolVarP(&insecure, "insecure", "i", false, "Skip TLS verification for S3/Minio custom endpoint")
	addCmd.Flags().StringVarP(&bucketName, "bucket", "b", "", "The name of the AWS S3 bucket to use for the cache")
	addCmd.Flags().StringVarP(&projectId, "projectid", "p", "", "The GCP project ID")
	addCmd.Flags().StringVarP(&storageAccount, "storageacc", "s", "", "The Azure storage account name of the container")
	addCmd.Flags().StringVarP(&containerName, "container", "c", "", "The Azure container name to use for the cache")
	addCmd.MarkFlagsRequiredTogether("storageacc", "container")
	// Tedious check to ensure we don't include arguments from different providers
	addCmd.MarkFlagsMutuallyExclusive("region", "storageacc")
	addCmd.MarkFlagsMutuallyExclusive("region", "container")
	addCmd.MarkFlagsMutuallyExclusive("bucket", "storageacc")
	addCmd.MarkFlagsMutuallyExclusive("bucket", "container")
	addCmd.MarkFlagsMutuallyExclusive("projectid", "storageacc")
	addCmd.MarkFlagsMutuallyExclusive("projectid", "container")
}
