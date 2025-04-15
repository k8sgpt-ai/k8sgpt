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
	"github.com/spf13/viper"
)

// AIClientFactory is an interface for creating AI clients
type AIClientFactory interface {
	NewClient(provider string) IAI
}

// DefaultAIClientFactory is the default implementation of AIClientFactory
type DefaultAIClientFactory struct{}

// NewClient creates a new AI client using the default implementation
func (f *DefaultAIClientFactory) NewClient(provider string) IAI {
	return NewClient(provider)
}

// ConfigProvider is an interface for accessing configuration
type ConfigProvider interface {
	UnmarshalKey(key string, rawVal interface{}) error
}

// ViperConfigProvider is the default implementation of ConfigProvider using Viper
type ViperConfigProvider struct{}

// UnmarshalKey unmarshals a key from the configuration using Viper
func (p *ViperConfigProvider) UnmarshalKey(key string, rawVal interface{}) error {
	return viper.UnmarshalKey(key, rawVal)
}

// Default instances to be used
var (
	DefaultClientFactory = &DefaultAIClientFactory{}
	DefaultConfigProvider = &ViperConfigProvider{}
)

// For testing - these variables can be overridden in tests
var (
	testAIClientFactory AIClientFactory = nil
	testConfigProvider  ConfigProvider  = nil
)

// GetAIClientFactory returns the test factory if set, otherwise the default
func GetAIClientFactory() AIClientFactory {
	if testAIClientFactory != nil {
		return testAIClientFactory
	}
	return DefaultClientFactory
}

// GetConfigProvider returns the test provider if set, otherwise the default
func GetConfigProvider() ConfigProvider {
	if testConfigProvider != nil {
		return testConfigProvider
	}
	return DefaultConfigProvider
}

// For testing - set the test implementations
func SetTestAIClientFactory(factory AIClientFactory) {
	testAIClientFactory = factory
}

func SetTestConfigProvider(provider ConfigProvider) {
	testConfigProvider = provider
}

// Reset test implementations
func ResetTestImplementations() {
	testAIClientFactory = nil
	testConfigProvider = nil
}