package server

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

func TestServe(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	s := &Config{
		Port:       "50059",
		Logger:     logger,
		EnableHttp: false,
	}

	go func() {
		err := s.Serve()
		time.Sleep(time.Second * 2)
		assert.NoError(t, err, "Serve should not return an error")
	}()

	// Wait until the server is ready to accept connections
	err := waitForPort("localhost:50059", 10*time.Second)
	assert.NoError(t, err, "Server should start without error")

	conn, err := grpc.Dial("localhost:50059", grpc.WithInsecure())
	assert.NoError(t, err, "Should be able to dial the server")
	defer conn.Close()

	// Test a simple gRPC reflection request
	cli := grpc_reflection_v1alpha.NewServerReflectionClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := cli.ServerReflectionInfo(ctx)
	assert.NoError(t, err, "Should be able to get server reflection info")
	assert.NotNil(t, resp, "Response should not be nil")

	// Cleanup
	err = s.Shutdown()
	assert.NoError(t, err, "Shutdown should not return an error")
}

func waitForPort(address string, timeout time.Duration) error {
	start := time.Now()
	for {
		conn, err := net.Dial("tcp", address)
		if err == nil {
			conn.Close()
			return nil
		}
		if time.Since(start) > timeout {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
}
