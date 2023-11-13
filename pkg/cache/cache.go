package cache

import (
	"fmt"
	"reflect"

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
	Configure(cacheInfo CacheProvider, noCache bool) error
	Store(key string, data string) error
	Load(key string) (string, error)
	List() ([]CacheObjectDetails, error)
	Exists(key string) bool
	IsCacheDisabled() bool
	GetName() string
}

func New(noCache bool, cacheType string) ICache {
	for _, t := range types {
		if cacheType == t.GetName() {
			return t
		}
	}
	return &FileBasedCache{}
}

// CacheProvider is the configuration for the cache provider when using a remote cache

func ParseCacheConfiguration() (CacheProvider, error) {
	var cacheInfo CacheProvider
	err := viper.UnmarshalKey("cache", &cacheInfo)
	if err != nil {
		return cacheInfo, err
	}
	return cacheInfo, nil
}

func NewCacheProvider(cacheType, bucketname, region, storageAccount, containerName, projectId string, noCache bool) (CacheProvider, error) {
	cache := New(false, cacheType)
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
		return CacheProvider{}, status.Error(codes.Internal, fmt.Sprintf("%s is not a possible option", cacheType))
	}

	err := cache.Configure(cProvider, noCache)
	if err != nil {
		return CacheProvider{}, err
	}
	return cProvider, nil
}

// If we have set a remote cache, return the remote cache type
func GetCacheConfiguration(noCache bool) (ICache, error) {
	// load remote cache if it is configured
	var cache ICache
	cacheInfo, err := ParseCacheConfiguration()
	if err != nil {
		return nil, err
	}

	switch {
	case !reflect.DeepEqual(cacheInfo.GCS, GCSCacheConfiguration{}):
		cache = &GCSCache{}
	case !reflect.DeepEqual(cacheInfo.Azure, AzureCacheConfiguration{}):
		cache = &AzureCache{}
	case !reflect.DeepEqual(cacheInfo.S3, S3CacheConfiguration{}):
		cache = &S3Cache{}
	default:
		cache = &FileBasedCache{}
	}

	cache.Configure(cacheInfo, noCache)

	return cache, nil
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
	return nil
	// var cacheInfo CacheProvider
	// err := viper.UnmarshalKey("cache", &cacheInfo)
	// if err != nil {
	// 	return status.Error(codes.Internal, "cache unmarshal")
	// }
	// if cacheInfo.BucketName == "" && cacheInfo.ContainerName == "" && cacheInfo.StorageAccount == "" {
	// 	return status.Error(codes.Internal, "no remote cache configured")
	// }

	// cacheInfo = CacheProvider{}
	// viper.Set("cache", cacheInfo)
	// err = viper.WriteConfig()
	// if err != nil {
	// 	return status.Error(codes.Internal, "unable to write config")
	// }

	// return nil

}
