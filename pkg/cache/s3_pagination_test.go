package cache

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func TestS3CacheListPaginates(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/xml")
		if r.URL.Query().Get("continuation-token") == "next-token" {
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>cache</Name>
  <IsTruncated>false</IsTruncated>
  <Contents><Key>second</Key><LastModified>2026-08-27T00:00:00.000Z</LastModified></Contents>
</ListBucketResult>`))
			return
		}
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>cache</Name>
  <IsTruncated>true</IsTruncated>
  <NextContinuationToken>next-token</NextContinuationToken>
  <Contents><Key>first</Key><LastModified>2026-08-27T00:00:00.000Z</LastModified></Contents>
</ListBucketResult>`))
	}))
	defer server.Close()

	sess := session.Must(session.NewSession(&aws.Config{
		Region:           aws.String("us-east-1"),
		Endpoint:         aws.String(server.URL),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials("key", "secret", ""),
	}))

	cache := &S3Cache{bucketName: "cache", session: s3.New(sess)}
	objects, err := cache.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(objects) != 2 || objects[0].Name != "first" || objects[1].Name != "second" {
		t.Fatalf("expected both pages from a truncated listing, got objects=%#v requests=%d", objects, requests)
	}
	if requests != 2 {
		t.Fatalf("expected two list requests, got %d", requests)
	}
}
