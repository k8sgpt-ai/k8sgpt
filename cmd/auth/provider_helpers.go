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

package auth

import (
	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
)

func newAIProviderFromAuthFlags(name string) ai.AIProvider {
	return ai.AIProvider{
		Name:            name,
		Model:           model,
		Password:        password,
		BaseURL:         baseURL,
		EndpointName:    endpointName,
		Engine:          engine,
		Temperature:     temperature,
		ProviderRegion:  providerRegion,
		ProviderId:      providerId,
		CompartmentId:   compartmentId,
		TopP:            topP,
		TopK:            topK,
		MaxTokens:       maxTokens,
		StopSequences:   stopSequences,
		OrganizationId:  organizationId,
		AzureAPIType:    azureAPIType,
		AzureAPIVersion: azureAPIVersion,
	}
}

func applyAIProviderUpdates(provider *ai.AIProvider, name string) {
	if name != "" {
		provider.Name = name
		color.Blue("Backend name updated successfully")
	}
	if model != "" {
		provider.Model = model
		color.Blue("Model updated successfully")
	}
	if password != "" {
		provider.Password = password
		color.Blue("Password updated successfully")
	}
	if baseURL != "" {
		provider.BaseURL = baseURL
		color.Blue("Base URL updated successfully")
	}
	if engine != "" {
		provider.Engine = engine
	}
	if organizationId != "" {
		provider.OrganizationId = organizationId
		color.Blue("Organization Id updated successfully")
	}
	if azureAPIType != "" {
		provider.AzureAPIType = azureAPIType
		color.Blue("AzureAPIType updated successfully")
	}
	if azureAPIVersion != "" {
		provider.AzureAPIVersion = azureAPIVersion
		color.Blue("AzureAPIVersion updated successfully")
	}
	provider.Temperature = temperature
}
