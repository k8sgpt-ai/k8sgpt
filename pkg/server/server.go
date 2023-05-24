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
	json "encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	rpc "buf.build/gen/go/k8sgpt-ai/k8sgpt/grpc/go/schema/v1/schemav1grpc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	Port           string
	MetricsPort    string
	Backend        string
	Key            string
	Token          string
	Output         string
	maxConcurrency int
	Handler        *handler
	Logger         *zap.Logger
	metricsServer  *http.Server
}

type Health struct {
	Status  string `json:"status"`
	Success int    `json:"success"`
	Failure int    `json:"failure"`
}

var health = Health{
	Status:  "ok",
	Success: 0,
	Failure: 0,
}

func (s *Config) Serve() error {

	var lis net.Listener
	var err error
	address := fmt.Sprintf(":%s", s.Port)
	lis, err = net.Listen("tcp", address)
	if err != nil {
		return err
	}
	s.Logger.Info(fmt.Sprintf("binding api to %s", s.Port))
	grpcServerUnaryInterceptor := grpc.UnaryInterceptor(logInterceptor(s.Logger))
	grpcServer := grpc.NewServer(grpcServerUnaryInterceptor)
	reflection.Register(grpcServer)
	rpc.RegisterServerServiceServer(grpcServer, s.Handler)
	if err := grpcServer.Serve(
		lis,
	); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
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

func (s *Config) healthzHandler(w http.ResponseWriter, r *http.Request) {
	js, err := json.MarshalIndent(health, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(js))
}

func getBoolParam(param string) bool {
	b, err := strconv.ParseBool(strings.ToLower(param))
	if err != nil {
		// Handle error if conversion fails
		return false
	}
	return b
}
