package query

import (
	schemav1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
	"context"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
)

func (h *Handler) Query(ctx context.Context, i *schemav1.QueryRequest) (
	*schemav1.QueryResponse,
	error,
) {
	aiClient := ai.NewClient(i.Backend)
	defer aiClient.Close()

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
