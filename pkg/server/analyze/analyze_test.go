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

package analyze

import (
	"context"
	"net"
	"testing"

	rpc "buf.build/gen/go/k8sgpt-ai/k8sgpt/grpc/go/schema/v1/schemav1grpc"
	schemav1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
	"google.golang.org/grpc"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

// fakeCustomAnalyzer serves the custom-analyzer gRPC API and returns whatever it is given.
type fakeCustomAnalyzer struct {
	rpc.UnimplementedCustomAnalyzerServiceServer
	resp *schemav1.RunResponse
}

func (f *fakeCustomAnalyzer) Run(context.Context, *schemav1.RunRequest) (*schemav1.RunResponse, error) {
	return f.resp, nil
}

// serveFakeAnalyzer starts one on a free port and returns its host and port.
func serveFakeAnalyzer(t *testing.T, resp *schemav1.RunResponse) (string, string) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	rpc.RegisterCustomAnalyzerServiceServer(srv, &fakeCustomAnalyzer{resp: resp})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	host, port, err := net.SplitHostPort(lis.Addr().String())
	if err != nil {
		t.Fatalf("split host/port: %v", err)
	}
	return host, port
}

// patchKubeClient points analysis.NewAnalysis at an empty fake cluster, so the
// core analyzers observe an empty (healthy) cluster without touching a real one.
func patchKubeClient(t *testing.T) {
	t.Helper()
	patches := gomonkey.ApplyFunc(kubernetes.NewClient, func(kubecontext, kubeconfig string) (*kubernetes.Client, error) {
		return &kubernetes.Client{
			Client: fake.NewSimpleClientset(),
			Config: &rest.Config{Host: "fake-server"},
		}, nil
	})
	t.Cleanup(patches.Reset)
}

// The server-mode Analyze handler must run custom analyzers when they are
// configured: their findings reach the response with the analyzer name as Kind,
// the problem count reflects the reported errors, and no transport errors leak
// into the response error list. This is the custom-analyzer path of
// pkg/server/analyze, which had no direct tests at all.
func TestHandlerAnalyzeRecordsCustomAnalyzerResults(t *testing.T) {
	host, port := serveFakeAnalyzer(t, &schemav1.RunResponse{
		Result: &schemav1.Result{
			Name:  "SelectedAnalyzer",
			Error: []*schemav1.ErrorDetail{{Text: "something is wrong"}},
		},
	})
	patchKubeClient(t)
	viper.Set("verbose", false)
	viper.Set("custom_analyzers", []map[string]interface{}{
		{"name": "SelectedAnalyzer", "connection": map[string]interface{}{"url": host, "port": port}},
	})
	t.Cleanup(func() { viper.Set("custom_analyzers", []interface{}{}) })

	h := &Handler{}
	resp, err := h.Analyze(context.Background(), &schemav1.AnalyzeRequest{
		Output:         "json",
		MaxConcurrency: 1,
	})
	require.NoError(t, err)
	require.Nil(t, resp.Error)
	require.Empty(t, resp.Errors)
	require.Equal(t, "ProblemDetected", resp.Status)
	require.Equal(t, int32(1), resp.Problems)
	require.Len(t, resp.Results, 1)
	require.Equal(t, "SelectedAnalyzer", resp.Results[0].Name)
	require.Equal(t, "SelectedAnalyzer", resp.Results[0].Kind, "Kind should default to the analyzer name")
}

// Without custom analyzers the handler still performs the core pass. On an
// empty cluster that is healthy: status OK, zero problems, no errors. An empty
// Output must default to json and MaxConcurrency 0 must not deadlock or panic.
func TestHandlerAnalyzeEmptyClusterIsOK(t *testing.T) {
	patchKubeClient(t)
	viper.Set("verbose", false)
	viper.Set("custom_analyzers", []interface{}{})

	h := &Handler{}
	resp, err := h.Analyze(context.Background(), &schemav1.AnalyzeRequest{})
	require.NoError(t, err)
	require.Nil(t, resp.Error)
	require.Empty(t, resp.Errors)
	require.Equal(t, "OK", resp.Status)
	require.Equal(t, int32(0), resp.Problems)
	require.Empty(t, resp.Results)
}
