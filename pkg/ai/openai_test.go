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

	gomock "github.com/golang/mock/gomock"
)

func TestOpenAI(t *testing.T) {

	mockController := gomock.NewController(t)
	defer mockController.Finish()

	mockConfig := NewMockIAIConfig(mockController)
	mockConfig.EXPECT().GetPassword().Return("testPassword").AnyTimes()
	mockConfig.EXPECT().GetModel().Return("testModel").AnyTimes()
	mockConfig.EXPECT().GetBaseURL().Return("testBaseURL").AnyTimes()

	openAIClient := &OpenAIClient{}

	err := openAIClient.Configure(mockConfig, "testLanguage")
	if err != nil {
		t.Errorf("Error configuring OpenAIClient: %v", err)
	}

	// @Aisuko - Need to mock c.client.CreateChatCompletion return value.
	// Otherwise it will always throw error with test data

	// _, err = openAIClient.GetCompletion(ctx, "testPrompt")
	// if err != nil {
	// 	t.Errorf("Error getting completion: %v", err)
	// }

}
