package cache

import "time"

type CacheProvider struct {
	GCS   GCSCacheConfiguration   `mapstructucre:"gcs" yaml:"gcs,omitempty"`
	Azure AzureCacheConfiguration `mapstructucre:"azure" yaml:"azure,omitempty"`
	S3    S3CacheConfiguration    `mapstructucre:"s3" yaml:"s3,omitempty"`
}

type CacheObjectDetails struct {
	Name      string
	UpdatedAt time.Time
}
