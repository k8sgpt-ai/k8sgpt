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

package ai

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"github.com/stretchr/testify/require"
)

func TestGoogleGenAIClientConfigureEmptyPassword(t *testing.T) {
	client := &GoogleGenAIClient{}

	var err error
	require.NotPanics(t, func() {
		err = client.Configure(&AIProvider{Name: googleAIClientName, Password: ""})
	})
	require.Error(t, err)
}

func TestGoogleGenAIClientConfigureAPIKey(t *testing.T) {
	client := &GoogleGenAIClient{}

	require.NoError(t, client.Configure(&AIProvider{Name: googleAIClientName, Password: "test-key"}))
	client.Close()
}

func TestGoogleGenAIClientConfigureInvalidCredentialsJSON(t *testing.T) {
	client := &GoogleGenAIClient{}

	require.Error(t, client.Configure(&AIProvider{Name: googleAIClientName, Password: "{"}))
}

func TestGoogleGenAIClientCitationWithoutURI(t *testing.T) {
	for _, tt := range []struct {
		name    string
		sources string
		want    string
	}{
		{"mixed citations", `[{"license":"MIT"},{"uri":"https://example.com/source","license":"Apache-2.0"}]`, "answer\nCitations:\n* https://example.com/source, Apache-2.0\n"},
		{"only citation without URI", `[{}]`, "answer\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, `{"candidates":[{"content":{"parts":[{"text":"answer"}],"role":"model"},"finishReason":"STOP","citationMetadata":{"citationSources":`+tt.sources+`}}]}`)
			}))
			defer server.Close()
			ctx := context.Background()
			sdk, err := genai.NewClient(ctx, option.WithAPIKey("test-key"), option.WithEndpoint(server.URL), option.WithHTTPClient(server.Client()))
			require.NoError(t, err)
			client := &GoogleGenAIClient{client: sdk, model: "gemini-pro", maxTokens: 100}
			defer client.Close()
			output, err := client.GetCompletion(ctx, "test")
			require.NoError(t, err)
			require.Equal(t, tt.want, output)
		})
	}
}
