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
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	gtwapi "sigs.k8s.io/gateway-api/apis/v1"
)

// The Gateway API analyzers (GatewayClass, Gateway, HTTPRoute) run concurrently
// (see analysis.RunAnalysis with max_concurrency). Each used to call
// gtwapi.AddToScheme(client.Scheme()) at the start of Analyze, mutating the shared
// runtime.Scheme maps concurrently and triggering a fatal "concurrent map writes".
//
// ensureGatewayScheme serializes the registration and performs it at most once per
// scheme, so after the first call there are no further writes to race against.
var (
	gatewaySchemeMu   sync.Mutex
	gatewaySchemeDone = map[*runtime.Scheme]struct{}{}
)

func ensureGatewayScheme(scheme *runtime.Scheme) error {
	gatewaySchemeMu.Lock()
	defer gatewaySchemeMu.Unlock()
	if _, ok := gatewaySchemeDone[scheme]; ok {
		return nil
	}
	if err := gtwapi.AddToScheme(scheme); err != nil {
		return err
	}
	gatewaySchemeDone[scheme] = struct{}{}
	return nil
}
