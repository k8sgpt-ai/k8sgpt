package cache

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
)

// Generate ICache implementation
type AzureCache struct {
	ctx           context.Context
	noCache       bool
	containerName string
	session       *azblob.Client
}

type AzureCacheConfiguration struct {
	StorageAccount string `mapstructure:"storageaccount" yaml:"storageaccount,omitempty"`
	ContainerName  string `mapstructure:"container" yaml:"container,omitempty"`
}

func (s *AzureCache) Configure(cacheInfo CacheProvider) error {
	s.ctx = context.Background()
	if cacheInfo.Azure.ContainerName == "" {
		log.Fatal("Azure Container name not configured")
	}
	if cacheInfo.Azure.StorageAccount == "" {
		log.Fatal("Azure Storage account not configured")
	}

	// We assume that Storage account is already in place
	blobUrl := fmt.Sprintf("https://%s.blob.core.windows.net/", cacheInfo.Azure.StorageAccount)
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}
	client, err := azblob.NewClient(blobUrl, credential, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Try to create the blob container
	_, err = client.CreateContainer(s.ctx, cacheInfo.Azure.ContainerName, nil)
	if err != nil {
		// TODO: Maybe there is a better way to check this?
		// docs: https://pkg.go.dev/github.com/Azure/azure-storage-blob-go/azblob
		if strings.Contains(err.Error(), "ContainerAlreadyExists") {
			// do nothing
		} else {
			return err
		}
	}
	s.containerName = cacheInfo.Azure.ContainerName
	s.session = client

	return nil

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

func (s *AzureCache) List() ([]CacheObjectDetails, error) {
	// List the files in the blob containerName
	files := []CacheObjectDetails{}

	pager := s.session.NewListBlobsFlatPager(s.containerName, &azblob.ListBlobsFlatOptions{
		Include: azblob.ListBlobsInclude{Snapshots: false, Versions: false},
	})

	for pager.More() {
		resp, err := pager.NextPage(s.ctx)
		if err != nil {
			return nil, err
		}

		for _, blob := range resp.Segment.BlobItems {
			files = append(files, CacheObjectDetails{
				Name:      *blob.Name,
				UpdatedAt: *blob.Properties.LastModified,
			})
		}
	}

	return files, nil
}

func (s *AzureCache) Remove(key string) error {
	_, err := s.session.DeleteBlob(s.ctx, s.containerName, key, &blob.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
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

func (s *AzureCache) GetName() string {
	return "azure"
}

func (s *AzureCache) DisableCache() {
	s.noCache = true
}
