package server

import (
	"context"

	schemav1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
	"github.com/k8sgpt-ai/k8sgpt/pkg/cache"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *handler) AddConfig(ctx context.Context, i *schemav1.AddConfigRequest) (*schemav1.AddConfigResponse, error,
) {

	resp, err := h.syncIntegration(ctx, i)
	if err != nil {
		return resp, err
	}

	if i.Cache != nil {
		var err error
		var remoteCache cache.CacheProvider

		switch i.Cache.GetCacheType().(type) {
		case *schemav1.Cache_AzureCache:
			remoteCache, err = cache.NewCacheProvider("azure", "", "", i.Cache.GetAzureCache().StorageAccount, i.Cache.GetAzureCache().ContainerName, "")
		case *schemav1.Cache_S3Cache:
			remoteCache, err = cache.NewCacheProvider("s3", i.Cache.GetS3Cache().BucketName, i.Cache.GetS3Cache().Region, "", "", "")
		case *schemav1.Cache_GcsCache:
			remoteCache, err = cache.NewCacheProvider("gcs", i.Cache.GetGcsCache().BucketName, i.Cache.GetGcsCache().Region, "", "", i.Cache.GetGcsCache().GetProjectId())
		default:
			return resp, status.Error(codes.InvalidArgument, "Invalid cache configuration")
		}

		if err != nil {
			return resp, err
		}
		err = cache.AddRemoteCache(remoteCache)
		if err != nil {
			return resp, err
		}

	}
	return resp, nil
}

func (h *handler) RemoveConfig(ctx context.Context, i *schemav1.RemoveConfigRequest) (*schemav1.RemoveConfigResponse, error,
) {
	err := cache.RemoveRemoteCache()
	if err != nil {
		return &schemav1.RemoveConfigResponse{}, err
	}

	// Remove any integrations is a TBD as it would be nice to make this more granular
	// Currently integrations can be removed in the AddConfig sync

	return &schemav1.RemoveConfigResponse{
		Status: "Successfully removed the remote cache",
	}, nil
}
