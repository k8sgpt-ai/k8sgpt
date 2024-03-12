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

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
)

func TestSliceContainsString(t *testing.T) {
	tests := []struct {
		name        string
		kubeContext string
		kubeConfig  string
		expectedErr string
	}{
		{
			name:        "empty config and empty context",
			kubeContext: "",
			kubeConfig:  "",
			expectedErr: "invalid configuration: no configuration has been provided, try setting KUBERNETES_MASTER environment variable",
		},
		{
			name:        "non empty config and empty context",
			kubeContext: "",
			kubeConfig:  "kube-config",
			expectedErr: "stat kube-config: no such file or directory",
		},
		{
			name:        "empty config and non empty context",
			kubeContext: "some-context",
			kubeConfig:  "",
			expectedErr: "context \"some-context\" does not exist",
		},
		{
			name:        "non empty config and non empty context",
			kubeContext: "minikube",
			kubeConfig:  "./testdata/kubeconfig",
			expectedErr: "Get \"https://192.168.49.2:8443/version\": dial tcp 192.168.49.2:8443",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.kubeContext, tt.kubeConfig)
			if tt.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.expectedErr)
				require.Nil(t, client)
			}
		})
	}
}

func TestKubernetesClient(t *testing.T) {
	client := Client{
		Config: &rest.Config{
			Host: "host",
		},
	}

	require.NotEmpty(t, client.GetConfig())
	require.Nil(t, client.GetClient())
	require.Nil(t, client.GetCtrlClient())
}
