package cache

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/spf13/viper"
)

type ICache interface {
	Store(key string, data string) error
	Load(key string) (string, error)
	List() ([]string, error)
	Exists(key string) bool
	IsCacheDisabled() bool
}

func New(noCache bool, remoteCache string) ICache {
	switch remoteCache {
	case "s3":
		return NewS3Cache(noCache)
	case "azure":
		return NewAzureCache(noCache)
	default:
		return &FileBasedCache{
			noCache: noCache,
		}
	}
}

// CacheProvider is the configuration for the cache provider when using a remote cache
type CacheProvider struct {
	BucketName     string `mapstructure:"bucketname" yaml:"bucketname,omitempty"`
	Region         string `mapstructure:"region" yaml:"region,omitempty"`
	StorageAccount string `mapstructure:"storageaccount" yaml:"storageaccount,omitempty"`
	ContainerName  string `mapstructure:"container" yaml:"container,omitempty"`
}

// NewCacheProvider constructs a new cache struct
func NewCacheProvider(bucketname, region, storageaccount, containername string) CacheProvider {
	return CacheProvider{
		BucketName:     bucketname,
		Region:         region,
		StorageAccount: storageaccount,
		ContainerName:  containername,
	}
}

// If we have set a remote cache, return the remote cache type
func RemoteCacheEnabled() (string, error) {
	// load remote cache if it is configured
	var cache CacheProvider
	err := viper.UnmarshalKey("cache", &cache)
	if err != nil {
		return "", err
	}
	if cache.BucketName != "" && cache.Region != "" {
		return "s3", nil
	} else if cache.StorageAccount != "" && cache.ContainerName != "" {
		return "azure", nil
	}
	return "", nil
}

func AddRemoteCache(cache CacheProvider) error {
	var cacheInfo CacheProvider
	err := viper.UnmarshalKey("cache", &cacheInfo)
	if err != nil {
		return err
	}

	cacheInfo.BucketName = cache.BucketName
	cacheInfo.Region = cache.Region
	cacheInfo.StorageAccount = cache.StorageAccount
	cacheInfo.ContainerName = cache.ContainerName
	viper.Set("cache", cacheInfo)
	err = viper.WriteConfig()
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
	if cacheInfo.BucketName == "" && cacheInfo.ContainerName == "" && cacheInfo.StorageAccount == "" {
		return status.Error(codes.Internal, "no remote cache configured")
	}

	cacheInfo = CacheProvider{}
	viper.Set("cache", cacheInfo)
	err = viper.WriteConfig()
	if err != nil {
		return status.Error(codes.Internal, "unable to write config")
	}

	return nil

}
