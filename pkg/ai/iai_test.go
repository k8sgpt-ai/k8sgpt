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
	"sync"
	"testing"
)

func TestAIProviderGetAzureAPIVersion(t *testing.T) {
	provider := AIProvider{AzureAPIVersion: "2024-02-15-preview"}

	if got := provider.GetAzureAPIVersion(); got != "2024-02-15-preview" {
		t.Fatalf("expected Azure API version to be returned, got %q", got)
	}
}

// TestNewClientReturnsFreshInstance ensures NewClient hands back a new client on
// every call rather than a shared package-level singleton. A shared instance
// would let concurrent callers overwrite each other's configuration.
func TestNewClientReturnsFreshInstance(t *testing.T) {
	first := NewClient(openAIClientName)
	second := NewClient(openAIClientName)

	if first == second {
		t.Fatalf("expected distinct clients per call, got the same instance %p", first)
	}
}

// TestNewClientConcurrentConfigure runs concurrent Configure calls on clients
// obtained from NewClient. Under the race detector this fails if NewClient
// returns a shared instance whose fields are mutated by Configure. Run with:
// go test -race ./pkg/ai/ -run TestNewClientConcurrentConfigure
func TestNewClientConcurrentConfigure(t *testing.T) {
	const goroutines = 8
	const iterations = 50

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			provider := &AIProvider{
				Model:       "model-" + string(rune('A'+id)),
				Password:    "token-" + string(rune('A'+id)),
				Temperature: float32(id),
			}
			for j := 0; j < iterations; j++ {
				c := NewClient(openAIClientName)
				if err := c.Configure(provider); err != nil {
					t.Errorf("Configure failed: %v", err)
					return
				}
			}
		}(i)
	}
	wg.Wait()
}
