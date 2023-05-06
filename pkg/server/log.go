package server

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

func logInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Call the handler to execute the gRPC request
		response, err := handler(ctx, req)

		duration := time.Since(start).Milliseconds()
		fields := []zap.Field{
			zap.Int64("duration_ms", duration),
			zap.String("method", info.FullMethod),
			zap.Any("request", req),
		}
		// Get the remote address from the context
		peer, ok := peer.FromContext(ctx)
		if ok {
			fields = append(fields, zap.String("remote_addr", peer.Addr.String()))
		}

		if err != nil {
			fields = append(fields, zap.Int32("status_code", int32(grpc.Code(err))))
		}
		message := "request completed"
		if err != nil {
			message = "request failed"
		}
		logRequest(logger, fields, int(grpc.Code(err)), message)

		return response, err
	}
}

func logRequest(logger *zap.Logger, fields []zap.Field, statusCode int, message string) {
	if statusCode >= 400 {
		logger.Error(message, fields...)
	} else {
		logger.Info(message, fields...)
	}
}
