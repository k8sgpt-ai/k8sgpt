/*
Copyright 2024 The K8sGPT Authors.
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

package integration

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestAnalyzerByIntegration(t *testing.T) {
	integration := NewIntegration()
	_, err := integration.Get("invalid-name")
	require.ErrorContains(t, err, "integration not found")

	tests := []struct {
		name         string
		expectedName string
		expectedErr  string
	}{
		{
			name:        "random",
			expectedErr: "analyzerbyintegration: no matches found",
		},
		{
			name:         "PrometheusConfigValidate",
			expectedName: "prometheus",
		},
		{
			name:         "PrometheusConfigRelabelReport",
			expectedName: "prometheus",
		},
		{
			name:         "VulnerabilityReport",
			expectedName: "trivy",
		},
		{
			name:         "ConfigAuditReport",
			expectedName: "trivy",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			name, err := integration.AnalyzerByIntegration(tt.name)
			if tt.expectedErr == "" {
				require.NoError(t, err)
				require.Equal(t, tt.expectedName, name)
			} else {
				require.ErrorContains(t, err, tt.expectedErr)
				require.Empty(t, name)
			}
		})
	}
}

func TestActivate(t *testing.T) {
	integration := NewIntegration()
	err := integration.Activate("prometheus", "", []string{}, true)
	require.ErrorContains(t, err, "error writing config file:")

	err = integration.Deactivate("prometheus", "")
	require.ErrorContains(t, err, "error writing config file:")

	configFileName := "config.json"
	_, err = os.CreateTemp("", configFileName)
	require.NoError(t, err)
	defer os.Remove(configFileName)

	// Set the configuration file in viper
	viper.SetConfigType("json")
	viper.SetConfigFile(configFileName)

	inteNotFoundErr := "integration not found"
	tests := []struct {
		name                    string
		namespace               string
		activeFilters           []string
		skipInstall             bool
		expectedIsActivate      bool
		expectedActivationErr   string
		expectedIsActivateError string
		expectedDeactivationErr string
	}{
		{
			name:                    "invalid integration",
			expectedActivationErr:   inteNotFoundErr,
			expectedIsActivateError: inteNotFoundErr,
			expectedDeactivationErr: inteNotFoundErr,
		},
		{
			name:               "prometheus",
			skipInstall:        true,
			expectedIsActivate: true,
		},
		{
			name:                    "trivy",
			skipInstall:             false,
			expectedActivationErr:   "failed to deploy trivy integration:",
			expectedDeactivationErr: "failed to undeploy trivy integration:",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := integration.Activate(tt.name, tt.namespace, tt.activeFilters, tt.skipInstall)
			if tt.expectedActivationErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.expectedActivationErr)
			}

			ok, err := integration.IsActivate(tt.name)
			if tt.expectedIsActivateError == "" {
				require.NoError(t, err)
				require.Equal(t, tt.expectedIsActivate, ok)
			} else {
				require.ErrorContains(t, err, tt.expectedIsActivateError)
			}

			err = integration.Deactivate(tt.name, tt.namespace)
			if tt.expectedDeactivationErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.expectedDeactivationErr)
			}
		})
	}
}
