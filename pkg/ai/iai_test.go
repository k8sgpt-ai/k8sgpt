/*
Copyright 2023 K8sGPT Contributors.

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
	"testing"
)

func TestAI(t *testing.T) {
	testProvider := &AIProvider{
		Name:     "testAIProvider",
		Model:    "testModel",
		Password: "testPassword",
		BaseURL:  "testBaseURL",
		Engine:   "testEngine",
	}
	// Checking GetBaseURL
	if testProvider.GetBaseURL() != "testBaseURL" {
		t.Errorf("Expected BaseURL to be testBaseURL, got %s", testProvider.GetBaseURL())
	}
	// Checking GetEngine
	if testProvider.GetEngine() != "testEngine" {
		t.Errorf("Expected Engine to be testEngine, got %s", testProvider.GetEngine())
	}
	// Checking GetModel
	if testProvider.GetModel() != "testModel" {
		t.Errorf("Expected Model to be testModel, got %s", testProvider.GetModel())
	}
	// Checking GetPassword
	if testProvider.GetPassword() != "testPassword" {
		t.Errorf("Expected Password to be testPassword, got %s", testProvider.GetPassword())
	}
	// Checking NeedPassword
	if NeedPassword(Backends[1]) == true {
		t.Errorf("Expected NeedPassword to be true, got false")
	}

	// Checking NewClient
	aiClient := NewClient(Backends[1])
	if aiClient.GetName() != "localai" {
		t.Errorf("Expected name to be localai, got %s", aiClient.GetName())
	}
	// Checking NewClient return default client
	defaultClient := NewClient("test")
	if defaultClient.GetName() != "openai" {
		t.Errorf("Expected name to be openai, got %s", defaultClient.GetName())
	}
}
