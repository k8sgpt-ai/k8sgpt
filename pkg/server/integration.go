package server

import (
	schemav1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
	"context"
	"fmt"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/integration"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	trivyName = "trivy"
)

// syncIntegration is aware of the following events
// A new integration added
// An integration removed from the Integration block
func (h *handler) syncIntegration(ctx context.Context,
	i *schemav1.AddConfigRequest) (*schemav1.AddConfigResponse, error,
) {
	response := &schemav1.AddConfigResponse{}
	integrationProvider := integration.NewIntegration()
	if i.Integrations == nil {
		// If there are locally activate integrations, disable them
		err := h.deactivateAllIntegrations(integrationProvider)
		if err != nil {
			return response, status.Error(codes.NotFound, "deactivation error")
		}
		return response, nil
	}
	coreFilters, _, _ := analyzer.ListFilters()
	// Update filters
	activeFilters := viper.GetStringSlice("active_filters")
	if len(activeFilters) == 0 {
		activeFilters = coreFilters
	}
	var err error = status.Error(codes.OK, "")
	if err != nil {
		fmt.Println(err)
	}
	deactivateFunc := func(integrationRef integration.IIntegration) error {
		namespace, err := integrationRef.GetNamespace()
		if err != nil {
			return err
		}
		err = integrationProvider.Deactivate(trivyName, namespace)
		if err != nil {
			return status.Error(codes.NotFound, "integration already deactivated")
		}
		return nil
	}
	integrationRef, err := integrationProvider.Get(trivyName)
	if err != nil {
		return response, status.Error(codes.NotFound, "provider get failure")
	}
	if i.Integrations.Trivy != nil {
		switch i.Integrations.Trivy.Enabled {
		case true:
			if b, err := integrationProvider.IsActivate(trivyName); err != nil {
				return response, status.Error(codes.Internal, "integration activation error")
			} else {
				if !b {
					err := integrationProvider.Activate(trivyName, i.Integrations.Trivy.Namespace,
						activeFilters, i.Integrations.Trivy.SkipInstall)
					if err != nil {
						return nil, err
					}
				} else {
					return response, status.Error(codes.AlreadyExists, "integration already active")
				}
			}
		case false:
			err = deactivateFunc(integrationRef)
			if err != nil {
				return nil, err
			}
			// This break is included purely for static analysis to pass
		}
	} else {
		// If Trivy has been removed, disable it
		err = deactivateFunc(integrationRef)
		if err != nil {
			return nil, err
		}
	}

	return response, err
}

func (*handler) ListIntegrations(ctx context.Context, req *schemav1.ListIntegrationsRequest) (*schemav1.ListIntegrationsResponse, error) {

	integrationProvider := integration.NewIntegration()
	// Update the requester with the status of Trivy
	trivy, err := integrationProvider.Get(trivyName)
	active := trivy.IsActivate()
	var skipInstall bool
	var namespace string = ""
	if active {
		namespace, err = trivy.GetNamespace()
		if err != nil {
			return nil, status.Error(codes.NotFound, "namespace not found")
		}
		if namespace == "" {
			skipInstall = true
		}
	}

	if err != nil {
		return nil, status.Error(codes.NotFound, "trivy integration")
	}
	resp := &schemav1.ListIntegrationsResponse{
		Trivy: &schemav1.Trivy{
			Enabled:     active,
			Namespace:   namespace,
			SkipInstall: skipInstall,
		},
	}

	return resp, nil
}

func (*handler) deactivateAllIntegrations(integrationProvider *integration.Integration) error {
	integrations := integrationProvider.List()
	for _, i := range integrations {
		b, _ := integrationProvider.IsActivate(i)
		if b {
			in, err := integrationProvider.Get(i)
			if err != nil {
				return err
			}
			namespace, err := in.GetNamespace()
			if err != nil {
				return err
			}
			if err == nil {
				if namespace != "" {
					err := integrationProvider.Deactivate(i, namespace)
					if err != nil {
						return err
					}
				} else {
					fmt.Printf("Skipping deactivation of %s, not installed\n", i)
				}
			} else {
				return err
			}
		}
	}
	return nil
}
