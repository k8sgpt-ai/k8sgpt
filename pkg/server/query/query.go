package query

import (
	"context"
	"fmt"

	schemav1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
)

func (h *Handler) Query(ctx context.Context, i *schemav1.QueryRequest) (
	*schemav1.QueryResponse,
	error,
) {
	// Create client factory and config provider
	factory := ai.GetAIClientFactory()
	configProvider := ai.GetConfigProvider()

	// Use the factory to create the client
	aiClient := factory.NewClient(i.Backend)
	defer aiClient.Close()

	var configAI ai.AIConfiguration
	if err := configProvider.UnmarshalKey("ai", &configAI); err != nil {
		return &schemav1.QueryResponse{
			Response: "",
			Error: &schemav1.QueryError{
				Message: fmt.Sprintf("Failed to unmarshal AI configuration: %v", err),
			},
		}, nil
	}

	var aiProvider ai.AIProvider
	for _, provider := range configAI.Providers {
		if i.Backend == provider.Name {
			aiProvider = provider
			break
		}
	}
	if aiProvider.Name == "" {
		return &schemav1.QueryResponse{
			Response: "",
			Error: &schemav1.QueryError{
				Message: fmt.Sprintf("AI provider %s not found in configuration", i.Backend),
			},
		}, nil
	}

	// Configure the AI client
	if err := aiClient.Configure(&aiProvider); err != nil {
		return &schemav1.QueryResponse{
			Response: "",
			Error: &schemav1.QueryError{
				Message: fmt.Sprintf("Failed to configure AI client: %v", err),
			},
		}, nil
	}

	resp, err := aiClient.GetCompletion(ctx, i.Query)
	var errMessage string = ""
	if err != nil {
		errMessage = err.Error()
	}
	return &schemav1.QueryResponse{
		Response: resp,
		Error: &schemav1.QueryError{
			Message: errMessage,
		},
	}, nil
}
