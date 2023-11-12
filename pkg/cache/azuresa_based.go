package cache

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/spf13/viper"
)

// Generate ICache implementation
type AzureCache struct {
	ctx           context.Context
	noCache       bool
	containerName string
	session       *azblob.Client
}

func (s *AzureCache) Store(key string, data string) error {
	// Store the object as a new file in the Azure blob storage with data as the content
	cacheData := []byte(data)
	_, err := s.session.UploadBuffer(s.ctx, s.containerName, key, cacheData, &azblob.UploadBufferOptions{})
	return err
}

func (s *AzureCache) Load(key string) (string, error) {
	// Load blob file contents
	load, err := s.session.DownloadStream(s.ctx, s.containerName, key, nil)
	if err != nil {
		return "", err
	}
	data := bytes.Buffer{}
	retryReader := load.NewRetryReader(s.ctx, &azblob.RetryReaderOptions{})
	_, err = data.ReadFrom(retryReader)
	if err != nil {
		return "", err
	}
	if err := retryReader.Close(); err != nil {
		return "", err
	}
	return data.String(), nil
}

func (s *AzureCache) List() ([]string, error) {
	// List the files in the blob containerName
	files := []string{}

	pager := s.session.NewListBlobsFlatPager(s.containerName, &azblob.ListBlobsFlatOptions{
		Include: azblob.ListBlobsInclude{Snapshots: false, Versions: false},
	})

	for pager.More() {
		resp, err := pager.NextPage(s.ctx)
		if err != nil {
			return nil, err
		}

		for _, blob := range resp.Segment.BlobItems {
			files = append(files, *blob.Name)
		}
	}

	return files, nil
}

func (s *AzureCache) Exists(key string) bool {
	// Check if the object exists in the blob storage
	pager := s.session.NewListBlobsFlatPager(s.containerName, &azblob.ListBlobsFlatOptions{
		Include: azblob.ListBlobsInclude{Snapshots: false, Versions: false},
	})

	for pager.More() {
		resp, err := pager.NextPage(s.ctx)
		if err != nil {
			return false
		}

		for _, blob := range resp.Segment.BlobItems {
			if *blob.Name == key {
				return true
			}
		}
	}

	return false
}

func (s *AzureCache) IsCacheDisabled() bool {
	return s.noCache
}

func NewAzureCache(nocache bool) ICache {
	ctx := context.Background()
	var cache CacheProvider
	err := viper.UnmarshalKey("cache", &cache)
	if err != nil {
		panic(err)
	}
	if cache.ContainerName == "" {
		log.Fatal("Azure Container name not configured")
	}
	if cache.StorageAccount == "" {
		log.Fatal("Azure Storage account not configured")
	}

	// We assume that Storage account is already in place
	blobUrl := fmt.Sprintf("https://%s.blob.core.windows.net/", cache.StorageAccount)
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}
	client, err := azblob.NewClient(blobUrl, credential, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Try to create the blob container
	_, err = client.CreateContainer(ctx, cache.ContainerName, nil)
	if err != nil {
		// TODO: Maybe there is a better way to check this?
		// docs: https://pkg.go.dev/github.com/Azure/azure-storage-blob-go/azblob
		if strings.Contains(err.Error(), "ContainerAlreadyExists") {
			// do nothing
		} else {
			log.Fatal(err)
		}
	}

	return &AzureCache{
		ctx:           ctx,
		noCache:       nocache,
		containerName: cache.ContainerName,
		session:       client,
	}
}
