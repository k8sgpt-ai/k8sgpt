package config

import (
	rpc "buf.build/gen/go/k8sgpt-ai/k8sgpt/grpc/go/schema/v1/schemav1grpc"
	schemav1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	rpc.UnimplementedServerConfigServiceServer
	// ShutdownFunc, when set, is invoked to gracefully stop the server in
	// response to a Shutdown RPC. It is injected by the server package,
	// which owns the listener lifecycle.
	ShutdownFunc func() error
}

func (h *Handler) Shutdown(ctx context.Context, request *schemav1.ShutdownRequest) (*schemav1.ShutdownResponse, error) {
	if h.ShutdownFunc == nil {
		return nil, status.Error(codes.Unimplemented, "shutdown is not supported by this server instance")
	}
	// Run asynchronously: shutting down closes the listener the current RPC
	// is being served on, so it must not block the response to this call.
	go func() {
		_ = h.ShutdownFunc()
	}()
	return &schemav1.ShutdownResponse{Status: "shutting down"}, nil
}
