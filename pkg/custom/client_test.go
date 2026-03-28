package custom

import (
	"context"
	"testing"

	schemav1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// mockAnalyzerClient implements rpc.CustomAnalyzerServiceClient for testing
type mockAnalyzerClient struct {
	resp *schemav1.RunResponse
	err  error
}

func (m *mockAnalyzerClient) Run(ctx context.Context, in *schemav1.RunRequest, opts ...grpc.CallOption) (*schemav1.RunResponse, error) {
	return m.resp, m.err
}

func TestClientRunMapsResponse(t *testing.T) {
	// prepare fake response
	resp := &schemav1.RunResponse{
		Result: &schemav1.Result{
			Name:         "AnalyzerA",
			Kind:         "Pod",
			Details:      "details",
			ParentObject: "Deployment/foo",
		},
	}
	cli := &Client{analyzerClient: &mockAnalyzerClient{resp: resp}}

	got, err := cli.Run()
	require.NoError(t, err)
	require.Equal(t, "AnalyzerA", got.Name)
	require.Equal(t, "Pod", got.Kind)
	require.Equal(t, "details", got.Details)
	require.Equal(t, "Deployment/foo", got.ParentObject)
	require.Len(t, got.Error, 0)
}
