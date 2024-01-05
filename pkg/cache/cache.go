package cache

import (
	"fmt"

	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	types = []ICache{
		&AzureCache{},
		&FileBasedCache{},
		&GCSCache{},
		&S3Cache{},
	}
)

type ICache interface {
	Configure(cacheInfo CacheProvider) error
	Store(key string, data string) error
	Load(key string) (string, error)
	List() ([]CacheObjectDetails, error)
	Remove(key string) error
	Exists(key string) bool
	IsCacheDisabled() bool
	GetName() string
	DisableCache()
}

func New(cacheType string) ICache {
	for _, t := range types {
		if cacheType == t.GetName() {
			return t
		}
	}
	return &FileBasedCache{}
}

func ParseCacheConfiguration() (CacheProvider, error) {
	var cacheInfo CacheProvider
	err := viper.UnmarshalKey("cache", &cacheInfo)
	if err != nil {
		return cacheInfo, err
	}
	return cacheInfo, nil
}

func NewCacheProvider(cacheType, bucketname, region, storageAccount, containerName, projectId string) (CacheProvider, error) {
	cProvider := CacheProvider{}

	switch {
	case cacheType == "azure":
		cProvider.Azure.ContainerName = containerName
		cProvider.Azure.StorageAccount = storageAccount
	case cacheType == "gcs":
		cProvider.GCS.BucketName = bucketname
		cProvider.GCS.ProjectId = projectId
		cProvider.GCS.Region = region
	case cacheType == "s3":
		cProvider.S3.BucketName = bucketname
		cProvider.S3.Region = region
	default:
		return CacheProvider{}, status.Error(codes.Internal, fmt.Sprintf("%s is not a valid option", cacheType))
	}

	cache := New(cacheType)
	err := cache.Configure(cProvider)
	if err != nil {
		return CacheProvider{}, err
	}
	return cProvider, nil
}

// If we have set a remote cache, return the remote cache configuration
func GetCacheConfiguration() (ICache, error) {
	cacheInfo, err := ParseCacheConfiguration()
	if err != nil {
		return nil, err
	}

	var cache ICache

	switch {
	case cacheInfo.GCS != GCSCacheConfiguration{}:
		cache = &GCSCache{}
	case cacheInfo.Azure != AzureCacheConfiguration{}:
		cache = &AzureCache{}
	case cacheInfo.S3 != S3CacheConfiguration{}:
		cache = &S3Cache{}
	default:
		cache = &FileBasedCache{}
	}

	err_config := cache.Configure(cacheInfo)

	return cache, err_config
}

func AddRemoteCache(cacheInfo CacheProvider) error {

	viper.Set("cache", cacheInfo)

	err := viper.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}

func RemoveRemoteCache() error {
	var cacheInfo CacheProvider
	err := viper.UnmarshalKey("cache", &cacheInfo)
	if err != nil {
		return status.Error(codes.Internal, "cache unmarshal")
	}

	cacheInfo = CacheProvider{}
	viper.Set("cache", cacheInfo)
	err = viper.WriteConfig()
	if err != nil {
		return status.Error(codes.Internal, "unable to write config")
	}

	return nil

}
