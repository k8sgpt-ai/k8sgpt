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

package kyverno

import (
	"context"
	"sync"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
)

// TestKyvernoAnalyzersConcurrentScheme guards against issue #1063: the
// PolicyReport and ClusterPolicyReport analyzers are registered as two separate
// analyzers and so run in parallel goroutines against one shared client. Both
// used to call v1alpha2.AddToScheme() from inside Analyze(), which writes the
// scheme's unguarded maps and crashed with "fatal error: concurrent map
// writes". The types are now registered once when the client is built.
//
// This covers the analyzer side only, against a client that is already built.
// The other side — building a client while an existing one is being analyzed —
// is covered by TestNewClientDoesNotShareSchemeState in pkg/kubernetes, where
// NewClient lives. Neither needs -race: Go's map implementation traps
// concurrent access on its own, so both fail under a plain `go test`.
func TestKyvernoAnalyzersConcurrentScheme(t *testing.T) {
	// The scheme is populated once, up front, the same way NewClient does it
	// when the real client is constructed.
	config := common.Analyzer{
		Client: &kubernetes.Client{
			CtrlClient: buildFakeClient(t),
		},
		Context:   context.Background(),
		Namespace: "test-ns",
	}

	analyzers := []common.IAnalyzer{
		KyvernoAnalyzer{policyReportAnalysis: true},
		KyvernoAnalyzer{clusterReportAnalysis: true},
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		for _, analyzer := range analyzers {
			wg.Add(1)
			go func(a common.IAnalyzer) {
				defer wg.Done()
				if _, err := a.Analyze(config); err != nil {
					t.Errorf("Analyze returned an unexpected error: %v", err)
				}
			}(analyzer)
		}
	}
	wg.Wait()
}
