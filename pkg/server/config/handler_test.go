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
	"testing"
	"time"

	schemav1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandler_Shutdown_Success(t *testing.T) {
	h := &Handler{}
	resp, err := h.Shutdown(context.Background(), &schemav1.ShutdownRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "Success", resp.Status)
}

func TestHandler_Shutdown_TriggersOnShutdown(t *testing.T) {
	done := make(chan struct{})
	h := &Handler{
		OnShutdown: func() {
			close(done)
		},
	}

	resp, err := h.Shutdown(context.Background(), &schemav1.ShutdownRequest{})
	require.NoError(t, err)
	assert.Equal(t, "Success", resp.Status)

	select {
	case <-done:
		// OnShutdown was successfully triggered
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for OnShutdown callback to be called")
	}
}

func TestHandler_Shutdown_NilRequest(t *testing.T) {
	h := &Handler{}
	resp, err := h.Shutdown(context.Background(), nil)
	require.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "request cannot be nil")
}

