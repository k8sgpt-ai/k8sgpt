package custom

import (
	"context"
	"net"
	"testing"

	rpc "buf.build/gen/go/k8sgpt-ai/k8sgpt/grpc/go/schema/v1/schemav1grpc"
	schemav1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	rpc.RegisterAnalyzerServiceServer(s, &fakeAnalyzerServiceServer{})
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
}

type fakeAnalyzerServiceServer struct {
	rpc.UnimplementedAnalyzerServiceServer
}

func (*fakeAnalyzerServiceServer) Run(ctx context.Context, req *schemav1.AnalyzerRunRequest) (*schemav1.AnalyzerRunResponse, error) {
	// Mock response
	return &schemav1.AnalyzerRunResponse{
		Result: &schemav1.Result{
			Name:         "test-name",
			Kind:         "test-kind",
			Details:      "test-details",
			ParentObject: "test-parent-object",
			Error: []*schemav1.ErrorDetail{
				{
					Text: "test-error-text",
				},
			},
		},
	}, nil
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func Test_NewClient(t *testing.T) {
	c := Connection{
		Url:  "bufnet",
		Port: "",
	}
	client, err := NewClient(c)
	assert.Nil(t, err, "Error should be nil")
	assert.Equal(t, "bufnet", c.Url, "Url should be bufnet")
	defer client.c.Close()
}

func Test_ClientRun(t *testing.T) {
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := rpc.NewAnalyzerServiceClient(conn)
	c := &Client{
		c:              conn,
		analyzerClient: client,
	}

	res, err := c.Run()
	assert.Equal(t, "test-name", res.Name, "Name should be test")
	assert.Equal(t, "test-kind", res.Kind, "Kind should be test")
	assert.Equal(t, "test-details", res.Details, "Details should be test")
	assert.Nil(t, err, "Error should be nil")

}
