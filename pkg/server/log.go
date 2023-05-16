// Copyright © 2023 K8sGPT.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
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
			fields = append(fields, zap.Int32("status_code", int32(status.Code(err))))
		}
		message := "request completed"
		if err != nil {
			message = fmt.Sprintf("request failed. %s", err.Error())
		}
		logRequest(logger, fields, int(status.Code(err)), message)

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
