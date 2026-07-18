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
	"math/rand"
	"net/http"
	"reflect"
	"sync"
	"testing"
)

type testCfg struct {
	password        string
	model           string
	baseURL         string
	proxyEndpoint   string
	endpointName    string
	engine          string
	temperature     float32
	providerRegion  string
	topP            float32
	topK            int32
	maxTokens       int
	stopSequences   []string
	providerID      string
	compartmentID   string
	organizationID  string
	azureAPIType    string
	azureAPIVersion string
	customHeaders   []http.Header
}

func (c *testCfg) GetPassword() string             { return c.password }
func (c *testCfg) GetModel() string                { return c.model }
func (c *testCfg) GetBaseURL() string               { return c.baseURL }
func (c *testCfg) GetProxyEndpoint() string         { return c.proxyEndpoint }
func (c *testCfg) GetEndpointName() string          { return c.endpointName }
func (c *testCfg) GetEngine() string                { return c.engine }
func (c *testCfg) GetTemperature() float32          { return c.temperature }
func (c *testCfg) GetProviderRegion() string        { return c.providerRegion }
func (c *testCfg) GetTopP() float32                 { return c.topP }
func (c *testCfg) GetTopK() int32                   { return c.topK }
func (c *testCfg) GetMaxTokens() int                { return c.maxTokens }
func (c *testCfg) GetStopSequences() []string       { return c.stopSequences }
func (c *testCfg) GetProviderId() string            { return c.providerID }
func (c *testCfg) GetCompartmentId() string         { return c.compartmentID }
func (c *testCfg) GetOrganizationId() string        { return c.organizationID }
func (c *testCfg) GetAzureAPIType() string          { return c.azureAPIType }
func (c *testCfg) GetAzureAPIVersion() string        { return c.azureAPIVersion }
func (c *testCfg) GetCustomHeaders() []http.Header  { return c.customHeaders }

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randPort() string {
	const digits = "0123456789"
	b := make([]byte, 4)
	for i := range b {
		b[i] = digits[rand.Intn(len(digits))]
	}
	return string(b)
}

// TestNewClient_ReturnsDistinctPointers ensures NewClient returns a fresh
// instance on every call instead of a shared package-level pointer.
func TestNewClient_ReturnsDistinctPointers(t *testing.T) {
	t.Parallel()
	c1 := NewClient("openai")
	c2 := NewClient("openai")

	if c1 == nil || c2 == nil {
		t.Fatalf("expected non-nil clients, got c1=%v c2=%v", c1, c2)
	}
	if c1 == c2 {
		t.Fatalf("expected distinct client instances, got identical pointers")
	}

	p1 := reflect.ValueOf(c1).Pointer()
	p2 := reflect.ValueOf(c2).Pointer()
	if p1 == p2 {
		t.Fatalf("expected different underlying pointers, got same: %v", p1)
	}
}

// TestConcurrent_NewClient_Configure_NoBleed drives NewClient + Configure from
// many goroutines at once and checks that no two goroutines ever observe the
// same client pointer, guarding against the shared-client data race this
// package's NewClient/Configure changes are meant to fix.
func TestConcurrent_NewClient_Configure_NoBleed(t *testing.T) {
	t.Parallel()

	const workers = 200
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(workers)

	type result struct {
		ptr uintptr
		err error
	}
	results := make(chan result, workers)

	for i := 0; i < workers; i++ {
		i := i
		go func() {
			defer wg.Done()
			<-start

			client := NewClient("openai")
			if client == nil {
				results <- result{0, nil}
				return
			}
			cfg := &testCfg{
				password:        "tok_" + randString(24) + "_" + randString(8),
				model:           "gpt-" + randString(6),
				baseURL:         "https://api.example.com/" + randString(10),
				proxyEndpoint:   "http://proxy.example.com:" + randPort(),
				endpointName:    "ep-" + randString(5),
				engine:          "eng-" + randString(5),
				temperature:     0.7,
				providerRegion:  "r-" + randString(4),
				topP:            0.9,
				topK:            int32(40 + i%5),
				maxTokens:       1024 + i%128,
				stopSequences:   []string{"</s>", "<|end|>"},
				providerID:      "pid-" + randString(6),
				compartmentID:   "cid-" + randString(6),
				organizationID:  "org-" + randString(6),
				azureAPIType:    "azure",
				azureAPIVersion: "2024-05-01",
				customHeaders:   []http.Header{{"X-Test-ID": {randString(12)}}},
			}
			err := client.Configure(cfg)
			results <- result{reflect.ValueOf(client).Pointer(), err}
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	seen := make(map[uintptr]struct{}, workers)
	var errs []error

	for r := range results {
		if r.ptr == 0 {
			t.Fatalf("received nil client instance in concurrency test")
		}
		if _, dup := seen[r.ptr]; dup {
			t.Fatalf("duplicate client pointer observed under concurrency: %v", r.ptr)
		}
		seen[r.ptr] = struct{}{}
		if r.err != nil {
			errs = append(errs, r.err)
		}
	}

	if len(seen) != workers {
		t.Fatalf("expected %d unique clients, saw %d", workers, len(seen))
	}
	if len(errs) > 0 {
		t.Fatalf("unexpected Configure errors under concurrency: count=%d, example=%v", len(errs), errs[0])
	}
}
