package cache

import (
	"bytes"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Generate ICache implementation
type S3Cache struct {
	noCache    bool
	bucketName string
	session    *s3.S3
}

type S3CacheConfiguration struct {
	Region     string `mapstructure:"region" yaml:"region,omitempty"`
	BucketName string `mapstructure:"bucketname" yaml:"bucketname,omitempty"`
}

func (s *S3Cache) Configure(cacheInfo CacheProvider) error {
	if cacheInfo.S3.BucketName == "" {
		log.Fatal("Bucket name not configured")
	}
	if cacheInfo.S3.Region == "" {
		log.Fatal("Region not configured")
	}
	s.bucketName = cacheInfo.S3.BucketName

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			Region: aws.String(cacheInfo.S3.Region),
		},
	}))

	s3Client := s3.New(sess)

	// Check if the bucket exists, if not create it
	_, err := s3Client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(cacheInfo.S3.BucketName),
	})
	if err != nil {
		_, err = s3Client.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(cacheInfo.S3.BucketName),
		})
		if err != nil {
			return err
		}
	}
	s.session = s3Client
	return nil
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

func (s *S3Cache) Remove(key string) error {
	_, err := s.session.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &s.bucketName,
		Key:    aws.String(key),
	})

	if err != nil {
		return err
	}
	return nil
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
	_, err_read := buf.ReadFrom(result.Body)
	result.Body.Close()
	return buf.String(), err_read
}

func (s *S3Cache) List() ([]CacheObjectDetails, error) {

	// List the files in the bucket
	result, err := s.session.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(s.bucketName)})
	if err != nil {
		return nil, err
	}

	var keys []CacheObjectDetails
	for _, item := range result.Contents {
		keys = append(keys, CacheObjectDetails{
			Name:      *item.Key,
			UpdatedAt: *item.LastModified,
		})
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

func (s *S3Cache) GetName() string {
	return "s3"
}

func (s *S3Cache) DisableCache() {
	s.noCache = true
}
