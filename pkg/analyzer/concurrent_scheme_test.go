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

package analyzer

import (
	"context"
	"sync"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gtwapi "sigs.k8s.io/gateway-api/apis/v1"
)

// TestGatewayAnalyzersConcurrentScheme guards against issue #1063: running the
// gateway-api analyzers in parallel used to crash with "fatal error:
// concurrent map writes" because each analyzer registered the gateway-api types
// into the shared client scheme from inside Analyze(). The types are now
// registered once when the client is built, so the analyzers can safely share a
// single client across goroutines. Run with -race to catch a regression.
func TestGatewayAnalyzersConcurrentScheme(t *testing.T) {
	// The gateway-api types are registered on the scheme once, up front, the
	// same way NewClient does it when the real client is constructed.
	testScheme := scheme.Scheme
	if err := gtwapi.Install(testScheme); err != nil {
		t.Fatal(err)
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(testScheme).Build()
	config := common.Analyzer{
		Client: &kubernetes.Client{
			CtrlClient: fakeClient,
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	analyzers := []common.IAnalyzer{
		GatewayAnalyzer{},
		GatewayClassAnalyzer{},
		HTTPRouteAnalyzer{},
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
