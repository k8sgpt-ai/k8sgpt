package server

import (
	"context"

	schemav1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
	"github.com/k8sgpt-ai/k8sgpt/pkg/integration"
)

func (*handler) ListIntegrations(ctx context.Context, req *schemav1.ListIntegrationsRequest) (*schemav1.ListIntegrationsResponse, error) {

	integrationProvider := integration.NewIntegration()
	integrations := integrationProvider.List()
	resp := &schemav1.ListIntegrationsResponse{
		Integrations: make([]string, 0),
	}
	for _, i := range integrations {
		b, _ := integrationProvider.IsActivate(i)
		if b {
			resp.Integrations = append(resp.Integrations, i)
		}
	}
	return resp, nil
}
