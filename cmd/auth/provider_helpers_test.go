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
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func TestAddCommandPersistsAzureAPIVersion(t *testing.T) {
	configureTestViper(t)
	resetAuthFlagState(t)

	setFlag(t, addCmd, "backend", "azureopenai")
	setFlag(t, addCmd, "model", "gpt-4o")
	setFlag(t, addCmd, "password", "token")
	setFlag(t, addCmd, "baseurl", "https://example.openai.azure.com")
	setFlag(t, addCmd, "engine", "deployment")
	setFlag(t, addCmd, "temperature", "0.7")
	setFlag(t, addCmd, "topp", "0.5")
	setFlag(t, addCmd, "topk", "50")
	setFlag(t, addCmd, "maxtokens", "2048")
	setFlag(t, addCmd, "azureAPIVersion", "2024-02-15-preview")

	addCmd.Run(addCmd, nil)

	var cfg ai.AIConfiguration
	if err := viper.UnmarshalKey("ai", &cfg); err != nil {
		t.Fatalf("failed to unmarshal ai config: %v", err)
	}
	if len(cfg.Providers) != 1 {
		t.Fatalf("expected one provider, got %d", len(cfg.Providers))
	}
	if cfg.Providers[0].AzureAPIVersion != "2024-02-15-preview" {
		t.Fatalf("expected Azure API version to be persisted, got %q", cfg.Providers[0].AzureAPIVersion)
	}
}

func TestUpdateCommandPersistsAzureAPIVersion(t *testing.T) {
	configureTestViper(t)
	resetAuthFlagState(t)

	viper.Set("ai", ai.AIConfiguration{
		Providers: []ai.AIProvider{{Name: "azureopenai", Model: "gpt-4o"}},
	})
	if err := viper.WriteConfig(); err != nil {
		t.Fatalf("failed to write initial config: %v", err)
	}

	setFlag(t, updateCmd, "backend", "azureopenai")
	setFlag(t, updateCmd, "temperature", "0.7")
	setFlag(t, updateCmd, "azureAPIVersion", "2024-06-01")

	updateCmd.Run(updateCmd, nil)

	var cfg ai.AIConfiguration
	if err := viper.UnmarshalKey("ai", &cfg); err != nil {
		t.Fatalf("failed to unmarshal ai config: %v", err)
	}
	if len(cfg.Providers) != 1 {
		t.Fatalf("expected one provider, got %d", len(cfg.Providers))
	}
	if cfg.Providers[0].AzureAPIVersion != "2024-06-01" {
		t.Fatalf("expected Azure API version to be updated, got %q", cfg.Providers[0].AzureAPIVersion)
	}
}

func TestNewAIProviderFromAuthFlagsIncludesAzureAPIVersion(t *testing.T) {
	resetAuthFlagState(t)

	model = "gpt-4o"
	password = "token"
	baseURL = "https://example.openai.azure.com"
	endpointName = "endpoint"
	engine = "deployment"
	temperature = 0.2
	providerRegion = "eastus"
	providerId = "provider-id"
	compartmentId = "compartment-id"
	topP = 0.8
	topK = 40
	maxTokens = 1024
	stopSequences = []string{"stop"}
	organizationId = "org-id"
	azureAPIType = "AZURE"
	azureAPIVersion = "2024-02-15-preview"

	provider := newAIProviderFromAuthFlags("azureopenai")

	if provider.Name != "azureopenai" {
		t.Fatalf("expected provider name azureopenai, got %q", provider.Name)
	}
	if provider.AzureAPIVersion != "2024-02-15-preview" {
		t.Fatalf("expected Azure API version to be preserved, got %q", provider.AzureAPIVersion)
	}
	if !reflect.DeepEqual(provider.StopSequences, []string{"stop"}) {
		t.Fatalf("expected stop sequences to be preserved, got %#v", provider.StopSequences)
	}
}

func TestApplyAIProviderUpdatesIncludesAzureAPIVersion(t *testing.T) {
	resetAuthFlagState(t)

	model = "gpt-4o-mini"
	password = "new-token"
	baseURL = "https://example.openai.azure.com"
	engine = "new-deployment"
	temperature = 0.4
	organizationId = "org-id"
	azureAPIType = "AZURE_AD"
	azureAPIVersion = "2024-06-01"

	provider := ai.AIProvider{Name: "azureopenai"}
	applyAIProviderUpdates(&provider, "azureopenai")

	if provider.Model != "gpt-4o-mini" {
		t.Fatalf("expected model to be updated, got %q", provider.Model)
	}
	if provider.AzureAPIVersion != "2024-06-01" {
		t.Fatalf("expected Azure API version to be updated, got %q", provider.AzureAPIVersion)
	}
	if provider.Temperature != 0.4 {
		t.Fatalf("expected temperature to be updated, got %v", provider.Temperature)
	}
}

func TestAddCommandModelDefaultPerBackend(t *testing.T) {
	tests := []struct {
		name          string
		backend       string
		expectedModel string
	}{
		{
			name:          "anthropic backend falls back to anthropic default model",
			backend:       "anthropic",
			expectedModel: anthropicDefaultModel,
		},
		{
			name:          "openai backend falls back to openai default model",
			backend:       "openai",
			expectedModel: defaultModel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configureTestViper(t)
			resetAuthFlagState(t)
			configAI = ai.AIConfiguration{}

			// Seed the model variable with the value cobra binds when --model is
			// omitted (the flag's registered default). This reproduces a real
			// `k8sgpt auth add -b <backend>` invocation without --model so the
			// per-backend fallback in add.go is exercised.
			model = addCmd.Flags().Lookup("model").DefValue

			setFlag(t, addCmd, "backend", tt.backend)
			setFlag(t, addCmd, "password", "token")
			setFlag(t, addCmd, "temperature", "0.7")
			setFlag(t, addCmd, "topp", "0.5")
			setFlag(t, addCmd, "topk", "50")
			setFlag(t, addCmd, "maxtokens", "2048")

			addCmd.Run(addCmd, nil)

			var cfg ai.AIConfiguration
			if err := viper.UnmarshalKey("ai", &cfg); err != nil {
				t.Fatalf("failed to unmarshal ai config: %v", err)
			}
			if len(cfg.Providers) != 1 {
				t.Fatalf("expected one provider, got %d", len(cfg.Providers))
			}
			if cfg.Providers[0].Model != tt.expectedModel {
				t.Fatalf("expected model %q for backend %q, got %q", tt.expectedModel, tt.backend, cfg.Providers[0].Model)
			}
		})
	}
}

func resetAuthFlagState(t *testing.T) {
	t.Helper()

	backend = ""
	password = ""
	baseURL = ""
	endpointName = ""
	model = ""
	engine = ""
	temperature = 0
	providerRegion = ""
	providerId = ""
	compartmentId = ""
	topP = 0
	topK = 0
	maxTokens = 0
	stopSequences = nil
	organizationId = ""
	azureAPIType = ""
	azureAPIVersion = ""
}

func configureTestViper(t *testing.T) {
	t.Helper()

	viper.Reset()
	configFile := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configFile, []byte("ai:\n  providers: []\n"), 0600); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("failed to read test config: %v", err)
	}
	t.Cleanup(viper.Reset)
}

func setFlag(t *testing.T, cmd interface{ Flags() *pflag.FlagSet }, name, value string) {
	t.Helper()

	if err := cmd.Flags().Set(name, value); err != nil {
		t.Fatalf("failed to set %s flag: %v", name, err)
	}
}
