package cache

import (
	"context"
	"io"
	"log"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type GCSCache struct {
	noCache    bool
	bucketName string
	projectId  string
	region     string
	session    *storage.Client
}

type GCSCacheConfiguration struct {
	ProjectId  string `mapstructure:"projectid" yaml:"projectid,omitempty"`
	Region     string `mapstructure:"region" yaml:"region,omitempty"`
	BucketName string `mapstructure:"bucketname" yaml:"bucketname,omitempty"`
}

func (s *GCSCache) Configure(cacheInfo CacheProvider, noCache bool) error {
	if cacheInfo.GCS.BucketName == "" {
		log.Fatal("Bucket name not configured")
	}
	if cacheInfo.GCS.Region == "" {
		log.Fatal("Region not configured")
	}
	if cacheInfo.GCS.ProjectId == "" {
		log.Fatal("ProjectID not configured")
	}
	s.bucketName = cacheInfo.GCS.BucketName
	s.projectId = cacheInfo.GCS.ProjectId
	s.region = cacheInfo.GCS.Region
	s.noCache = noCache
	ctx := context.Background()
	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	_, err = storageClient.Bucket(s.bucketName).Attrs(ctx)
	if err == storage.ErrBucketNotExist {
		storageClient.Bucket(s.bucketName).Create(ctx, s.projectId, &storage.BucketAttrs{
			Location: s.region,
		})
	}
	s.session = storageClient
	return nil
}

func (s *GCSCache) Store(key string, data string) error {
	ctx := context.Background()
	wc := s.session.Bucket(s.bucketName).Object(key).NewWriter(ctx)

	if _, err := wc.Write([]byte(data)); err != nil {
		return err
	}

	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}

func (s *GCSCache) Load(key string) (string, error) {
	ctx := context.Background()
	reader, err := s.session.Bucket(s.bucketName).Object(key).NewReader(ctx)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (s *GCSCache) List() ([]CacheObjectDetails, error) {
	ctx := context.Background()
	var files []CacheObjectDetails

	items := s.session.Bucket(s.bucketName).Objects(ctx, nil)
	for {
		attrs, err := items.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		files = append(files, CacheObjectDetails{
			Name:      attrs.Name,
			UpdatedAt: attrs.Updated,
		})
	}
	return files, nil
}

func (s *GCSCache) Exists(key string) bool {
	ctx := context.Background()
	obj := s.session.Bucket(s.bucketName).Object(key)
	_, err := obj.Attrs(ctx)
	return err == nil
}

func (s *GCSCache) IsCacheDisabled() bool {
	return s.noCache
}

func (s *GCSCache) GetName() string {
	return "gcs"
}
