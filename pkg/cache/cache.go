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

func New(noCache bool, remoteCache bool) ICache {
	if remoteCache {
		return NewS3Cache(noCache)
	}
	return &FileBasedCache{
		noCache: noCache,
	}
}

// CacheProvider is the configuration for the cache provider when using a remote cache
type CacheProvider struct {
	BucketName string `mapstructure:"bucketname"`
	Region     string `mapstructure:"region"`
}

func RemoteCacheEnabled() (bool, error) {
	// load remote cache if it is configured
	var cache CacheProvider
	err := viper.UnmarshalKey("cache", &cache)
	if err != nil {
		return false, err
	}
	if cache.BucketName != "" && cache.Region != "" {
		return true, nil
	}
	return false, nil
}

func AddRemoteCache(bucketName string, region string) error {
	var cacheInfo CacheProvider
	err := viper.UnmarshalKey("cache", &cacheInfo)
	if err != nil {
		return err
	}

	cacheInfo.BucketName = bucketName
	cacheInfo.Region = region
	viper.Set("cache", cacheInfo)
	err = viper.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}

func RemoveRemoteCache(bucketName string) error {
	var cacheInfo CacheProvider
	err := viper.UnmarshalKey("cache", &cacheInfo)
	if err != nil {
		return status.Error(codes.Internal, "cache unmarshal")
	}
	if cacheInfo.BucketName == "" {
		return status.Error(codes.Internal, "no cache configured")
	}

	cacheInfo = CacheProvider{}
	viper.Set("cache", cacheInfo)
	err = viper.WriteConfig()
	if err != nil {
		return status.Error(codes.Internal, "unable to write config")
	}

	return nil

}
