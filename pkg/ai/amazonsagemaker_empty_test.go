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

package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sagemakerruntime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSageMakerClient(t *testing.T, body string) *SageMakerAIClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Endpoint:    aws.String(srv.URL),
		Credentials: credentials.NewStaticCredentials("id", "secret", ""),
	}))
	return &SageMakerAIClient{client: sagemakerruntime.New(sess), endpoint: "test"}
}

func TestSageMakerGetCompletion_EmptyContent(t *testing.T) {
	for name, body := range map[string]string{
		"blank":   `[{"generation":{"role":"assistant","content":"   "}}]`,
		"missing": `[{"generation":{"role":"assistant"}}]`,
	} {
		t.Run(name, func(t *testing.T) {
			c := newTestSageMakerClient(t, body)
			out, err := c.GetCompletion(context.Background(), "hi")
			require.Error(t, err)
			assert.Empty(t, out)
		})
	}
}

func TestSageMakerGetCompletion_Content(t *testing.T) {
	c := newTestSageMakerClient(t, `[{"generation":{"role":"assistant","content":"ok"}}]`)
	out, err := c.GetCompletion(context.Background(), "hi")
	require.NoError(t, err)
	assert.Equal(t, "ok", out)
}
