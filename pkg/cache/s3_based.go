package cache

import (
	"bytes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/viper"
)

// Generate ICache implementation
type S3Cache struct {
	noCache    bool
	bucketName string
	session    *s3.S3
}

func (s *S3Cache) Store(key string, data string) error {
	// Store the object as a new file in the bucket with data as the content
	_, err := s.session.PutObject(&s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(bytes.NewReader([]byte(data))),
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	return err

}

func (s *S3Cache) Load(key string) (string, error) {

	// Retrieve the object from the bucket and load it into a string
	result, err := s.session.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(result.Body)
	result.Body.Close()
	return buf.String(), nil
}

func (s *S3Cache) List() ([]string, error) {

	// List the files in the bucket
	result, err := s.session.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(s.bucketName)})
	if err != nil {
		return nil, err
	}

	var keys []string
	for _, item := range result.Contents {
		keys = append(keys, *item.Key)
	}

	return keys, nil
}

func (s *S3Cache) Exists(key string) bool {
	// Check if the object exists in the bucket
	_, err := s.session.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	return err == nil

}

func (s *S3Cache) IsCacheDisabled() bool {
	return s.noCache
}

func NewS3Cache(nocache bool) ICache {

	var cache CacheProvider
	err := viper.UnmarshalKey("cache", &cache)
	if err != nil {
		panic(err)
	}
	if cache.BucketName == "" {
		panic("Bucket name not configured")
	}
	if cache.Region == "" {
		panic("Region not configured")
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			Region: aws.String(cache.Region),
		},
	}))

	s := s3.New(sess)

	// Check if the bucket exists, if not create it
	_, err = s.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(cache.BucketName),
	})
	if err != nil {
		_, err = s.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(cache.BucketName),
		})
		if err != nil {
			panic(err)
		}
	}

	return &S3Cache{
		noCache:    nocache,
		session:    s,
		bucketName: cache.BucketName,
	}
}
