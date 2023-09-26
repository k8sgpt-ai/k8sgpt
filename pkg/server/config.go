package server

import (
	schemav1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
	"context"
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
		// Remote cache
		if i.Cache.BucketName == "" || i.Cache.Region == "" {
			return resp, status.Error(codes.InvalidArgument, "cache arguments")
		}

		err := cache.AddRemoteCache(i.Cache.BucketName, i.Cache.Region)
		if err != nil {
			return resp, err
		}
	}
	return resp, nil
}

func (h *handler) RemoveConfig(ctx context.Context, i *schemav1.RemoveConfigRequest) (*schemav1.RemoveConfigResponse, error,
) {
	err := cache.RemoveRemoteCache(i.Cache.BucketName)
	if err != nil {
		return &schemav1.RemoveConfigResponse{}, err
	}

	// Remove any integrations is a TBD as it would be nice to make this more granular
	// Currently integrations can be removed in the AddConfig sync

	return &schemav1.RemoveConfigResponse{
		Status: "Successfully removed the remote cache",
	}, nil
}
