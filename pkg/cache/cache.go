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
		&InterplexCache{},
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

func NewCacheProvider(cacheType, bucketname, region, endpoint, storageAccount, containerName, projectId string, insecure bool) (CacheProvider, error) {
	cProvider := CacheProvider{}

	switch {
	case cacheType == "azure":
		cProvider.Azure.ContainerName = containerName
		cProvider.Azure.StorageAccount = storageAccount
		cProvider.CurrentCacheType = "azure"
	case cacheType == "gcs":
		cProvider.GCS.BucketName = bucketname
		cProvider.GCS.ProjectId = projectId
		cProvider.GCS.Region = region
		cProvider.CurrentCacheType = "gcs"
	case cacheType == "s3":
		cProvider.S3.BucketName = bucketname
		cProvider.S3.Region = region
		cProvider.S3.Endpoint = endpoint
		cProvider.S3.InsecureSkipVerify = insecure
		cProvider.CurrentCacheType = "s3"
	case cacheType == "interplex":
		cProvider.Interplex.ConnectionString = endpoint
		cProvider.CurrentCacheType = "interplex"
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
	case cacheInfo.CurrentCacheType == "gcs":
		cache = &GCSCache{}
	case cacheInfo.CurrentCacheType == "azure":
		cache = &AzureCache{}
	case cacheInfo.CurrentCacheType == "s3":
		cache = &S3Cache{}
	case cacheInfo.CurrentCacheType == "interplex":
		cache = &InterplexCache{}
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
