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
	"net/http"
)

// Mock AIProvider implementation for testing
type mockAIProvider struct {
	knowledgeBase  string
	model          string
	providerRegion string
	temperature    float32
	topP           float32
	maxTokens      int
}

func (p *mockAIProvider) GetKnowledgeBase() string {
	return p.knowledgeBase
}

func (p *mockAIProvider) GetModel() string {
	return p.model
}

func (p *mockAIProvider) GetProviderRegion() string {
	return p.providerRegion
}

func (p *mockAIProvider) GetTemperature() float32 {
	return p.temperature
}

func (p *mockAIProvider) GetTopP() float32 {
	return p.topP
}

func (p *mockAIProvider) GetMaxTokens() int {
	return p.maxTokens
}

// Implement other required methods from IAIConfig
func (p *mockAIProvider) GetPassword() string { return "" }
func (p *mockAIProvider) GetBaseURL() string { return "" }
func (p *mockAIProvider) GetProxyEndpoint() string { return "" }
func (p *mockAIProvider) GetEndpointName() string { return "" }
func (p *mockAIProvider) GetEngine() string { return "" }
func (p *mockAIProvider) GetTopK() int32 { return 0 }
func (p *mockAIProvider) GetProviderId() string { return "" }
func (p *mockAIProvider) GetCompartmentId() string { return "" }
func (p *mockAIProvider) GetOrganizationId() string { return "" }
func (p *mockAIProvider) GetCustomHeaders() []http.Header { return nil }
