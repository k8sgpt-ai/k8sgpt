package config

import (
	"context"

	rpc "buf.build/gen/go/k8sgpt-ai/k8sgpt/grpc/go/schema/v1/schemav1grpc"
	schemav1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	rpc.UnimplementedServerConfigServiceServer
	OnShutdown func()
}

func (h *Handler) Shutdown(ctx context.Context, request *schemav1.ShutdownRequest) (*schemav1.ShutdownResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if h.OnShutdown != nil {
		go h.OnShutdown()
	}
	return &schemav1.ShutdownResponse{
		Status: "Success",
	}, nil
}
