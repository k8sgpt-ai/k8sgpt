package cache

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stretchr/testify/assert"
)

type mockedS3Client struct {
	s3iface.S3API
	mu      sync.Mutex
	buckets map[string]bool
	files   map[string][]byte
	tags    map[string]map[string]string
}

func newMockedS3Client() *mockedS3Client {
	return &mockedS3Client{
		buckets: map[string]bool{},
		files:   map[string][]byte{},
		tags:    map[string]map[string]string{},
	}
}

func (m *mockedS3Client) HeadBucket(in *s3.HeadBucketInput) (*s3.HeadBucketOutput, error) {
	fmt.Println("HeadBucket")
	m.mu.Lock()
	defer m.mu.Unlock()

	bucket := *in.Bucket
	if _, ok := m.buckets[bucket]; !ok {
		return nil, awserr.New(s3.ErrCodeNoSuchBucket, "bucket does not exist", nil)
	}

	return &s3.HeadBucketOutput{}, nil
}

func (m *mockedS3Client) CreateBucket(in *s3.CreateBucketInput) (*s3.CreateBucketOutput, error) {
	fmt.Println("CreateBucket")
	m.mu.Lock()
	defer m.mu.Unlock()

	bucket := *in.Bucket
	if _, ok := m.buckets[bucket]; ok {
		return nil, awserr.New(s3.ErrCodeBucketAlreadyExists, "bucket already exists", nil)
	}

	m.buckets[bucket] = true // Add bucket to the map
	return &s3.CreateBucketOutput{}, nil
}

func (m *mockedS3Client) PutObject(in *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fmt.Println("PutObject")
	key := path.Join(*in.Bucket, *in.Key)
	m.files[key], _ = ioutil.ReadAll(in.Body)

	m.tags[key] = map[string]string{}
	if in.Tagging != nil {
		u, err := url.Parse("/?" + *in.Tagging)
		if err != nil {
			panic(fmt.Errorf("Unable to parse AWS S3 Tagging string %q: %w", *in.Tagging, err))
		}

		q := u.Query()
		for k := range q {
			m.tags[key][k] = q.Get(k)
		}
	}

	return &s3.PutObjectOutput{}, nil
}

func (m *mockedS3Client) GetObject(in *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fmt.Println("GetObject")
	key := path.Join(*in.Bucket, *in.Key)
	if _, ok := m.files[key]; !ok {
		return nil, awserr.New(s3.ErrCodeNoSuchKey, "key does not exist", nil)
	}

	return &s3.GetObjectOutput{
		Body: ioutil.NopCloser(bytes.NewReader(m.files[key])),
	}, nil
}

func (m *mockedS3Client) DeleteObject(in *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := *in.Key
	delete(m.files, key)
	delete(m.tags, key)

	return &s3.DeleteObjectOutput{}, nil
}

func (m *mockedS3Client) ListObjectsV2(input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var objects []*s3.Object
	for key := range m.files {
		objects = append(objects, &s3.Object{Key: aws.String(key), LastModified: aws.Time(time.Now())})
	}

	return &s3.ListObjectsV2Output{Contents: objects}, nil
}

func (m *mockedS3Client) HeadObject(input *s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := *input.Key
	if _, ok := m.files[key]; !ok {
		return nil, awserr.New(s3.ErrCodeNoSuchKey, "object does not exist", nil)
	}

	return &s3.HeadObjectOutput{}, nil
}

func TestS3Cache_Configure(t *testing.T) {
	// Mocked CacheProvider
	cacheInfo := CacheProvider{
		S3: S3CacheConfiguration{
			Region:     "us-west-1",
			BucketName: "test-bucket-k8sgpt",
		},
	}

	// Mocked S3 client
	mockS3Client := newMockedS3Client()

	// Create S3Cache instance
	s3Cache := &S3Cache{
		session:    mockS3Client,
		bucketName: "test-bucket-k8sgpt",
	}

	// Test Configure method
	err := s3Cache.Configure(cacheInfo)
	assert.NoError(t, err)

	// Verify that the session is set
	assert.NotNil(t, s3Cache.session)

	// Verify that the bucket name is set correctly
	assert.Equal(t, "test-bucket-k8sgpt", s3Cache.bucketName)
}

func TestS3Cache_Configure_CreateBucket(t *testing.T) {
	// Mocked CacheProvider
	cacheInfo := CacheProvider{
		S3: S3CacheConfiguration{
			Region:     "us-west-1",
			BucketName: "test-bucket-k8s",
		},
	}

	// Mocked S3 client
	mockS3Client := newMockedS3Client()

	// Create S3Cache instance
	s3Cache := &S3Cache{
		session: mockS3Client,
	}

	// Test Configure method
	err := s3Cache.Configure(cacheInfo)
	assert.NoError(t, err)

	// Verify that the session is set
	assert.NotNil(t, s3Cache.session)

	// Verify that the bucket name is set correctly
	assert.Equal(t, "test-bucket-k8s", s3Cache.bucketName)
}

func TestS3Cache_Store(t *testing.T) {
	// Create a mocked S3 client
	mockS3Client := newMockedS3Client()

	// Create S3Cache instance with the mocked client
	s3Cache := &S3Cache{
		session:    mockS3Client,
		bucketName: "test-bucket",
	}

	// Test Store method
	err := s3Cache.Store("test-key", "test-data")
	assert.NoError(t, err)

	// Verify that the object was stored
	_, err = mockS3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("test-bucket"),
		Key:    aws.String("test-key"),
	})
	assert.NoError(t, err)
}

func TestS3Cache_Load(t *testing.T) {
	// Create a mocked S3 client
	mockS3Client := newMockedS3Client()

	// Create S3Cache instance with the mocked client
	s3Cache := &S3Cache{
		session:    mockS3Client,
		bucketName: "test-bucket",
	}

	// Store a test object
	err := s3Cache.Store("test-key", "test-data")
	assert.NoError(t, err)

	// Test Load method
	data, err := s3Cache.Load("test-key")
	assert.NoError(t, err)
	assert.Equal(t, "test-data", data)
}

func TestS3Cache_List(t *testing.T) {
	// Create a mocked S3 client
	mockS3Client := newMockedS3Client()

	// Create S3Cache instance with the mocked client
	s3Cache := &S3Cache{
		session:    mockS3Client,
		bucketName: "test-bucket",
	}

	// Add some files to the mocked S3 client
	mockS3Client.files["file1.txt"] = []byte("file1 content")
	mockS3Client.files["file2.txt"] = []byte("file2 content")

	// Test List method
	keys, err := s3Cache.List()
	assert.NoError(t, err)
	assert.Len(t, keys, 2)
	assert.Equal(t, "file1.txt", keys[0].Name)
	assert.Equal(t, "file2.txt", keys[1].Name)
}

func TestS3Cache_Remove(t *testing.T) {
	// Create a mocked S3 client
	mockS3Client := newMockedS3Client()

	// Create S3Cache instance with the mocked client
	s3Cache := &S3Cache{
		session:    mockS3Client,
		bucketName: "test-bucket",
	}

	// Add a file to the mocked S3 client
	mockS3Client.files["test-key"] = []byte("test data")

	// Test Remove method
	err := s3Cache.Remove("test-key")
	assert.NoError(t, err)

	// Verify that the object was removed
	exists := s3Cache.Exists("test-key")
	assert.False(t, exists)
}

func TestS3Cache_Exists(t *testing.T) {
	// Create a mocked S3 client
	mockS3Client := newMockedS3Client()

	// Create S3Cache instance with the mocked client
	s3Cache := &S3Cache{
		session:    mockS3Client,
		bucketName: "test-bucket",
	}

	// Add a file to the mocked S3 client
	mockS3Client.files["test-key"] = []byte("test data")

	// Test Exists method
	exists := s3Cache.Exists("test-key")
	assert.True(t, exists)

	// Test with non-existing key
	exists = s3Cache.Exists("non-existing-key")
	assert.False(t, exists)
}

func TestS3Cache_IsCacheDisabled(t *testing.T) {
	// Create S3Cache instance
	s3Cache := &S3Cache{
		noCache: true,
	}

	// Test IsCacheDisabled method
	disabled := s3Cache.IsCacheDisabled()
	assert.True(t, disabled)
}

func TestS3Cache_GetName(t *testing.T) {
	// Create S3Cache instance
	s3Cache := &S3Cache{}

	// Test GetName method
	name := s3Cache.GetName()
	assert.Equal(t, "s3", name)
}

func TestS3Cache_DisableCache(t *testing.T) {
	// Create S3Cache instance
	s3Cache := &S3Cache{}

	// Test DisableCache method
	s3Cache.DisableCache()
	assert.True(t, s3Cache.IsCacheDisabled())
}
