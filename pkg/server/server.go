/*
Copyright 2023 The K8sGPT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	gw "buf.build/gen/go/k8sgpt-ai/k8sgpt/grpc-ecosystem/gateway/v2/schema/v1/server-service/schemav1gateway"
	rpc "buf.build/gen/go/k8sgpt-ai/k8sgpt/grpc/go/schema/v1/schemav1grpc"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	Port          string
	MetricsPort   string
	Backend       string
	Key           string
	Token         string
	Output        string
	Handler       *handler
	Logger        *zap.Logger
	metricsServer *http.Server
	listener      net.Listener
	EnableHttp    bool
}

type Health struct {
	Status  string `json:"status"`
	Success int    `json:"success"`
	Failure int    `json:"failure"`
}

//nolint:unused
var health = Health{
	Status:  "ok",
	Success: 0,
	Failure: 0,
}

func (s *Config) Shutdown() error {
	return s.listener.Close()
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func (s *Config) Serve() error {
	var lis net.Listener
	var err error
	address := fmt.Sprintf(":%s", s.Port)
	lis, err = net.Listen("tcp", address)
	if err != nil {
		return err
	}

	s.listener = lis
	s.Logger.Info(fmt.Sprintf("binding api to %s", s.Port))
	grpcServerUnaryInterceptor := grpc.UnaryInterceptor(logInterceptor(s.Logger))
	grpcServer := grpc.NewServer(grpcServerUnaryInterceptor)
	reflection.Register(grpcServer)
	rpc.RegisterServerServiceServer(grpcServer, s.Handler)

	if s.EnableHttp {
		s.Logger.Info("enabling rest/http api")
		gwmux := runtime.NewServeMux()
		err = gw.RegisterServerServiceHandlerFromEndpoint(context.Background(), gwmux, fmt.Sprintf("localhost:%s", s.Port), []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
		if err != nil {
			log.Fatalln("Failed to register gateway:", err)
		}

		srv := &http.Server{
			Addr:    address,
			Handler: h2c.NewHandler(grpcHandlerFunc(grpcServer, gwmux), &http2.Server{}),
		}

		if err := srv.Serve(lis); err != nil {
			return err
		}
	} else {
		if err := grpcServer.Serve(
			lis,
		); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	return nil
}

func (s *Config) ServeMetrics() error {
	s.Logger.Info(fmt.Sprintf("binding metrics to %s", s.MetricsPort))
	s.metricsServer = &http.Server{
		ReadHeaderTimeout: 3 * time.Second,
		Addr:              fmt.Sprintf(":%s", s.MetricsPort),
	}
	s.metricsServer.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/healthz":
			w.WriteHeader(http.StatusOK)
		case "/metrics":
			promhttp.Handler().ServeHTTP(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	if err := s.metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
