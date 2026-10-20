package cache

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/require"
)

func TestS3CacheListPaginates(t *testing.T) {
	const lastModified = "2026-08-27T00:00:00.000Z"
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		require.Equal(t, "/test-bucket", r.URL.Path)
		w.Header().Set("Content-Type", "application/xml")
		switch requests {
		case 1:
			require.Empty(t, r.URL.Query().Get("continuation-token"))
			_, _ = fmt.Fprintf(w, `<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><Name>test-bucket</Name><IsTruncated>true</IsTruncated><NextContinuationToken>next-token</NextContinuationToken><Contents><Key>first</Key><LastModified>%s</LastModified></Contents></ListBucketResult>`, lastModified)
		case 2:
			require.Equal(t, "next-token", r.URL.Query().Get("continuation-token"))
			_, _ = fmt.Fprintf(w, `<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><Name>test-bucket</Name><IsTruncated>false</IsTruncated><Contents><Key>second</Key><LastModified>%s</LastModified></Contents></ListBucketResult>`, lastModified)
		default:
			t.Fatalf("unexpected S3 list request %d", requests)
		}
	}))
	defer server.Close()

	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials("test", "test", ""),
		Endpoint:    aws.String(server.URL), Region: aws.String("us-east-1"), S3ForcePathStyle: aws.Bool(true),
	}))
	cache := &S3Cache{bucketName: "test-bucket", session: s3.New(sess)}
	got, err := cache.List()
	require.NoError(t, err)
	require.Equal(t, 2, requests)
	wantTime, err := time.Parse(time.RFC3339Nano, lastModified)
	require.NoError(t, err)
	require.Equal(t, []CacheObjectDetails{{Name: "first", UpdatedAt: wantTime}, {Name: "second", UpdatedAt: wantTime}}, got)
}
