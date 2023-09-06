package server

import (
	"context"
	"errors"

	schemav1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/cache"
	"github.com/k8sgpt-ai/k8sgpt/pkg/integration"
	"github.com/spf13/viper"
)

func (h *handler) AddConfig(ctx context.Context, i *schemav1.AddConfigRequest) (*schemav1.AddConfigResponse, error,
) {

	if i.Integrations != nil {
		coreFilters, _, _ := analyzer.ListFilters()
		// Update filters
		activeFilters := viper.GetStringSlice("active_filters")
		if len(activeFilters) == 0 {
			activeFilters = coreFilters
		}
		integration := integration.NewIntegration()

		if i.Integrations.Trivy != nil {
			// Enable/Disable Trivy
			var err = integration.Activate("trivy", i.Integrations.Trivy.Namespace,
				activeFilters, i.Integrations.Trivy.Enabled)
			return &schemav1.AddConfigResponse{
				Status: err.Error(),
			}, err
		}
	}
	if i.Cache != nil {
		// Remote cache
		if i.Cache.BucketName == "" || i.Cache.Region == "" {
			return &schemav1.AddConfigResponse{}, errors.New("BucketName & Region are required")
		}

		err := cache.AddRemoteCache(i.Cache.BucketName, i.Cache.Region)
		if err != nil {
			return &schemav1.AddConfigResponse{
				Status: err.Error(),
			}, err
		}
	}
	return &schemav1.AddConfigResponse{
		Status: "Configuration updated.",
	}, nil
}

func (h *handler) RemoveConfig(ctx context.Context, i *schemav1.RemoveConfigRequest) (*schemav1.RemoveConfigResponse, error,
) {
	err := cache.RemoveRemoteCache(i.Cache.BucketName)
	if err != nil {
		return &schemav1.RemoveConfigResponse{
			Status: err.Error(),
		}, err
	}

	return &schemav1.RemoveConfigResponse{
		Status: "Successfully removed the remote cache",
	}, nil
}
