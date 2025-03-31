package cache

import (
	rpc "buf.build/gen/go/interplex-ai/schemas/grpc/go/protobuf/schema/v1/schemav1grpc"
	schemav1 "buf.build/gen/go/interplex-ai/schemas/protocolbuffers/go/protobuf/schema/v1"
	"context"
	"errors"
	"google.golang.org/grpc"
	"net"
	"testing"
)

func TestInterplexCache(t *testing.T) {
	cache := &InterplexCache{
		configuration: InterplexCacheConfiguration{
			ConnectionString: "localhost:50051",
		},
	}

	// Mock GRPC server setup
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			t.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		rpc.RegisterCacheServiceServer(s, &mockCacheService{})
		if err := s.Serve(lis); err != nil {
			t.Fatalf("failed to serve: %v", err)
		}
	}()

	t.Run("TestStore", func(t *testing.T) {
		err := cache.Store("key1", "value1")
		if err != nil {
			t.Errorf("Error storing value: %v", err)
		}
	})

	t.Run("TestLoad", func(t *testing.T) {
		value, err := cache.Load("key1")
		if err != nil {
			t.Errorf("Error loading value: %v", err)
		}
		if value != "value1" {
			t.Errorf("Expected value1, got %v", value)
		}
	})

	t.Run("TestExists", func(t *testing.T) {
		exists := cache.Exists("key1")
		if !exists {
			t.Errorf("Expected key1 to exist")
		}
	})
}

type mockCacheService struct {
	rpc.UnimplementedCacheServiceServer
	data map[string]string
}

func (m *mockCacheService) Set(ctx context.Context, req *schemav1.SetRequest) (*schemav1.SetResponse, error) {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[req.Key] = req.Value
	return &schemav1.SetResponse{}, nil
}

func (m *mockCacheService) Get(ctx context.Context, req *schemav1.GetRequest) (*schemav1.GetResponse, error) {
	value, exists := m.data[req.Key]
	if !exists {
		return nil, errors.New("key not found")
	}
	return &schemav1.GetResponse{Value: value}, nil
}
