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
	"context"
	"testing"

	gomock "github.com/golang/mock/gomock"
	cache "github.com/k8sgpt-ai/k8sgpt/pkg/cache"
)

func TestNoOpAIClient(t *testing.T) {
	noOpAIClient := &NoOpAIClient{}
	// Checking GetName
	if noOpAIClient.GetName() != "noopai" {
		t.Errorf("Expected name to be noopai, got %s", noOpAIClient.GetName())
	}

	ctx := context.Background()
	mockController := gomock.NewController(t)
	defer mockController.Finish()
	mockConfig := NewMockIAIConfig(mockController)
	mockConfig.EXPECT().GetPassword().Return("testPassword").AnyTimes()
	mockConfig.EXPECT().GetModel().Return("testModel").AnyTimes()

	err := noOpAIClient.Configure(mockConfig, "testLanguage")
	if err != nil {
		t.Errorf("Error configuring NoOpAIClient: %v", err)
	}
	_, err = noOpAIClient.GetCompletion(ctx, "testPrompt")
	if err != nil {
		t.Errorf("Error getting completion: %v", err)
	}

	cache := cache.NewMockICache(mockController)
	// cache.EXPECT().Store("test", "test").Return(nil).AnyTimes()
	cache.EXPECT().Store("3684ab2e854cb0f8e79e13cc400ea1684fa4a4aa330cc49461db45a2d5109009", "SSBhbSBhIG5vb3AgcmVzcG9uc2UgdG8gdGhlIHByb21wdCB0ZXN0").Return(nil).AnyTimes()

	// @Aisuko - Here the parameter should be replaced by the base64 encoded string
	// cache.EXPECT().Load("test").Return("test", nil).AnyTimes()
	// checking Parse
	response, err := noOpAIClient.Parse(ctx, []string{"test"}, cache)
	if err != nil {
		t.Errorf("Error parsing: %v", err)
	}
	if response != "I am a noop response to the prompt test" {
		t.Errorf("Expected response to be I am a noop response to the prompt test, got %s", response)
	}
}
