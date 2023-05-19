package cache

import (
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
