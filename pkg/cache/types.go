package cache

import "time"

type CacheProvider struct {
	CurrentCacheType string                      `mapstructure:"currentCacheType" yaml:"currentCacheType"`
	GCS              GCSCacheConfiguration       `mapstructure:"gcs" yaml:"gcs,omitempty"`
	Azure            AzureCacheConfiguration     `mapstructure:"azure" yaml:"azure,omitempty"`
	S3               S3CacheConfiguration        `mapstructure:"s3" yaml:"s3,omitempty"`
	Interplex        InterplexCacheConfiguration `mapstructure:"interplex" yaml:"interplex,omitempty"`
}

type CacheObjectDetails struct {
	Name      string
	UpdatedAt time.Time
}
