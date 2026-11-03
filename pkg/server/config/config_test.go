/*
Copyright 2026 The K8sGPT Authors.
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

package config

import (
	"context"
	"path/filepath"
	"testing"

	schemav1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
	"github.com/k8sgpt-ai/k8sgpt/pkg/custom"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestApplyConfig_NilRequest(t *testing.T) {
	h := &Handler{}
	err := h.ApplyConfig(context.Background(), nil)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
	require.Contains(t, st.Message(), "request cannot be nil")
}

func TestApplyConfig_NilCustomAnalyzer(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.SetConfigFile(filepath.Join(t.TempDir(), "config.yaml"))

	h := &Handler{}
	req := &schemav1.AddConfigRequest{
		CustomAnalyzers: []*schemav1.CustomAnalyzer{nil},
	}
	err := h.ApplyConfig(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
	require.Contains(t, st.Message(), "custom analyzer cannot be nil")
}

func TestApplyConfig_EmptyCustomAnalyzerName(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.SetConfigFile(filepath.Join(t.TempDir(), "config.yaml"))

	h := &Handler{}
	req := &schemav1.AddConfigRequest{
		CustomAnalyzers: []*schemav1.CustomAnalyzer{
			{
				Name: "",
				Connection: &schemav1.Connection{
					Url:  "localhost",
					Port: "8080",
				},
			},
		},
	}
	err := h.ApplyConfig(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
	require.Contains(t, st.Message(), "custom analyzer name cannot be empty")
}

func TestApplyConfig_NilConnection(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.SetConfigFile(filepath.Join(t.TempDir(), "config.yaml"))

	h := &Handler{}
	req := &schemav1.AddConfigRequest{
		CustomAnalyzers: []*schemav1.CustomAnalyzer{
			{
				Name:       "test-analyzer",
				Connection: nil,
			},
		},
	}
	err := h.ApplyConfig(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
	require.Contains(t, st.Message(), "custom analyzer test-analyzer connection cannot be nil")
}

func TestApplyConfig_SuccessAndDeduplication(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.SetConfigFile(filepath.Join(t.TempDir(), "config.yaml"))

	h := &Handler{}
	req := &schemav1.AddConfigRequest{
		CustomAnalyzers: []*schemav1.CustomAnalyzer{
			{
				Name: "test-analyzer",
				Connection: &schemav1.Connection{
					Url:  "localhost",
					Port: "8080",
				},
			},
		},
	}

	err := h.ApplyConfig(context.Background(), req)
	require.NoError(t, err)

	var analyzers []custom.CustomAnalyzer
	err = viper.UnmarshalKey("custom_analyzers", &analyzers)
	require.NoError(t, err)
	require.Len(t, analyzers, 1)
	require.Equal(t, "test-analyzer", analyzers[0].Name)
	require.Equal(t, "localhost", analyzers[0].Connection.Url)
	require.Equal(t, "8080", analyzers[0].Connection.Port)

	// Apply again with duplicate name and a new analyzer
	req2 := &schemav1.AddConfigRequest{
		CustomAnalyzers: []*schemav1.CustomAnalyzer{
			{
				Name: "test-analyzer",
				Connection: &schemav1.Connection{
					Url:  "localhost",
					Port: "8080",
				},
			},
			{
				Name: "second-analyzer",
				Connection: &schemav1.Connection{
					Url:  "remotehost",
					Port: "9090",
				},
			},
		},
	}

	err = h.ApplyConfig(context.Background(), req2)
	require.NoError(t, err)

	analyzers = nil
	err = viper.UnmarshalKey("custom_analyzers", &analyzers)
	require.NoError(t, err)
	require.Len(t, analyzers, 2)
	require.Equal(t, "second-analyzer", analyzers[1].Name)
	require.Equal(t, "remotehost", analyzers[1].Connection.Url)
	require.Equal(t, "9090", analyzers[1].Connection.Port)
}

func TestAddConfig_NilRequest(t *testing.T) {
	h := &Handler{}
	resp, err := h.AddConfig(context.Background(), nil)
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestAddConfig_Success(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.SetConfigFile(filepath.Join(t.TempDir(), "config.yaml"))

	h := &Handler{}
	req := &schemav1.AddConfigRequest{
		CustomAnalyzers: []*schemav1.CustomAnalyzer{
			{
				Name: "my-custom-analyzer",
				Connection: &schemav1.Connection{
					Url:  "127.0.0.1",
					Port: "50051",
				},
			},
		},
	}

	resp, err := h.AddConfig(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	var analyzers []custom.CustomAnalyzer
	err = viper.UnmarshalKey("custom_analyzers", &analyzers)
	require.NoError(t, err)
	require.Len(t, analyzers, 1)
	require.Equal(t, "my-custom-analyzer", analyzers[0].Name)
}

func TestSyncIntegration_NilRequest(t *testing.T) {
	h := &Handler{}
	resp, err := h.syncIntegration(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestRemoveConfig_NilRequest(t *testing.T) {
	h := &Handler{}
	resp, err := h.RemoveConfig(context.Background(), nil)
	require.Error(t, err)
	require.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestRemoveConfig_Success(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	configFile := filepath.Join(t.TempDir(), "config.yaml")
	viper.SetConfigFile(configFile)
	require.NoError(t, viper.WriteConfig())

	h := &Handler{}
	resp, err := h.RemoveConfig(context.Background(), &schemav1.RemoveConfigRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Contains(t, resp.Status, "Successfully removed the remote cache")
}
