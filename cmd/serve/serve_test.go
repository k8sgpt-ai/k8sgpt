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

package serve

import (
	"net/http"
	"testing"
)

func TestProviderFromEnvIncludesAzureAPIVersion(t *testing.T) {
	t.Setenv("K8SGPT_BACKEND", "azureopenai")
	t.Setenv("K8SGPT_PASSWORD", "token")
	t.Setenv("K8SGPT_MODEL", "gpt-4o")
	t.Setenv("K8SGPT_BASEURL", "https://example.openai.azure.com")
	t.Setenv("K8SGPT_ENGINE", "deployment")
	t.Setenv("K8SGPT_AZURE_API_TYPE", "AZURE")
	t.Setenv("K8SGPT_AZURE_API_VERSION", "2024-02-15-preview")
	t.Setenv("K8SGPT_PROXY_ENDPOINT", "http://proxy.example.com")
	t.Setenv("K8SGPT_PROVIDER_ID", "provider-id")

	headers := http.Header{}
	headers.Set("x-test-header", "enabled")

	provider, ok := providerFromEnv(
		func() []http.Header { return []http.Header{headers} },
		func() float32 { return 0.2 },
		func() float32 { return 0.8 },
		func() int32 { return 40 },
		func() int { return 1024 },
	)

	if !ok {
		t.Fatal("expected provider to be created from environment")
	}
	if provider.AzureAPIVersion != "2024-02-15-preview" {
		t.Fatalf("expected Azure API version to be preserved, got %q", provider.AzureAPIVersion)
	}
	if provider.CustomHeaders[0].Get("x-test-header") != "enabled" {
		t.Fatalf("expected custom headers to be preserved, got %#v", provider.CustomHeaders)
	}
	if provider.TopK != 40 || provider.MaxTokens != 1024 {
		t.Fatalf("expected sampling settings to be preserved, got topK=%d maxTokens=%d", provider.TopK, provider.MaxTokens)
	}
}

func TestProviderFromEnvReturnsFalseWhenUnset(t *testing.T) {
	provider, ok := providerFromEnv(
		func() []http.Header { return nil },
		func() float32 { return 0 },
		func() float32 { return 0 },
		func() int32 { return 0 },
		func() int { return 0 },
	)

	if ok {
		t.Fatalf("expected no provider, got %#v", provider)
	}
}
