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

package analysis

import (
	"context"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/cache"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/stretchr/testify/require"
)

// recordingAIClient captures the prompt it is handed, so a test can assert on
// what would actually leave the machine for the provider.
type recordingAIClient struct {
	ai.NoOpAIClient
	prompt string
}

func (c *recordingAIClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	c.prompt = prompt
	return c.NoOpAIClient.GetCompletion(ctx, prompt)
}

func newDisabledCache() cache.ICache {
	c := cache.New("disabled-cache")
	c.DisableCache()
	return c
}

// TestGetAIResultsAnonymizesResourceIdentity covers issue #560: analyzers copy
// Kubernetes event messages straight into Failure.Text with an empty
// Failure.Sensitive, so under --anonymize the resource names inside those
// messages were still sent to the AI provider verbatim.
func TestGetAIResultsAnonymizesResourceIdentity(t *testing.T) {
	client := &recordingAIClient{}
	a := Analysis{
		AIClient: client,
		Cache:    newDisabledCache(),
		Results: []common.Result{
			{
				Kind: "Pod",
				Name: "prod-payments/checkout-worker-7d9f",
				Error: []common.Failure{
					{
						// The shape every event-derived failure has today.
						Text:      "Readiness probe failed for pod checkout-worker-7d9f in namespace prod-payments: connection refused",
						Sensitive: []common.Sensitive{},
					},
				},
			},
		},
	}

	require.NoError(t, a.GetAIResults("json", true))

	require.NotContains(t, client.prompt, "checkout-worker-7d9f",
		"pod name was sent to the AI provider despite --anonymize")
	require.NotContains(t, client.prompt, "prod-payments",
		"namespace was sent to the AI provider despite --anonymize")

	// Masking is reversed on the way back, so the user still reads real names.
	require.Contains(t, a.Results[0].Details, "checkout-worker-7d9f")
	require.Contains(t, a.Results[0].Details, "prod-payments")
}

// Analyzers that already declare their own Sensitive entries must keep working
// alongside the derived ones.
func TestGetAIResultsCombinesDeclaredAndDerivedSensitive(t *testing.T) {
	client := &recordingAIClient{}
	a := Analysis{
		AIClient: client,
		Cache:    newDisabledCache(),
		Results: []common.Result{
			{
				Kind: "StatefulSet",
				Name: "default/example",
				Error: []common.Failure{
					{
						Text: "pod example-0 in namespace default is not running, node worker.internal",
						Sensitive: []common.Sensitive{
							{Unmasked: "example-0", Masked: "MASKED-POD"},
						},
					},
				},
			},
		},
	}

	require.NoError(t, a.GetAIResults("json", true))

	require.NotContains(t, client.prompt, "example-0", "declared value leaked")
	require.NotContains(t, client.prompt, "default", "derived namespace leaked")

	require.Contains(t, a.Results[0].Details, "example-0")
	require.Contains(t, a.Results[0].Details, "default")
}

// A cluster-scoped result has a bare name with no "/" separator, and a name
// with a leading separator must not produce an empty mask: masking the empty
// string matches everywhere and would shred the text in both directions.
func TestResourceSensitive(t *testing.T) {
	for name, tc := range map[string]struct {
		resource string
		expected []string
	}{
		"namespaced":     {"prod/api-server", []string{"prod", "api-server"}},
		"cluster scoped": {"ip-10-0-1-2.ec2.internal", []string{"ip-10-0-1-2.ec2.internal"}},
		"container":      {"prod/api-server/sidecar", []string{"prod", "api-server", "sidecar"}},
		"empty segments": {"/api-server/", []string{"api-server"}},
		"empty name":     {"", nil},
	} {
		t.Run(name, func(t *testing.T) {
			sensitive := resourceSensitive(tc.resource)

			require.Len(t, sensitive, len(tc.expected))
			for i, want := range tc.expected {
				require.Equal(t, want, sensitive[i].Unmasked)
				require.NotEmpty(t, sensitive[i].Masked)
				require.NotEqual(t, want, sensitive[i].Masked)
			}
		})
	}
}
