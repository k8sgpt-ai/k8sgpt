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

package auth

import (
	"bytes"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestAddCmd(t *testing.T) {
	require.Equal(t, "add", addCmd.Name())

	// TODO: Create a temp config file
	configFileName := "config.json"
	_, err := os.CreateTemp("", configFileName)
	require.NoError(t, err)
	defer os.Remove(configFileName)

	// Set the configuration file in viper
	viper.SetConfigType("json")
	viper.SetConfigFile(configFileName)

	tests := []struct {
		name           string
		args           map[string]string
		expectedOutput string
	}{
		{
			name: "default backend",
			args: map[string]string{
				"password": "test-pass",
			},
			expectedOutput: "Warning: backend input is empty, will use the default value: openai\nWarning: model input is empty, will use the default value: gpt-3.5-turbo\nopenai added to the AI backend provider list\n",
		},
		{
			name: "localai backend",
			args: map[string]string{
				"backend": "localai",
			},
			expectedOutput: "localai added to the AI backend provider list\n",
		},
		{
			name: "amazonsagemaker backend",
			args: map[string]string{
				"backend": "amazonsagemaker",
			},
			expectedOutput: "amazonsagemaker added to the AI backend provider list\n",
		},
		{
			name: "amazonbedrock backend",
			args: map[string]string{
				"backend": "amazonbedrock",
			},
			expectedOutput: "amazonbedrock added to the AI backend provider list\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.args {
				oldVal := addCmd.Flag(key).Value.String()
				err := addCmd.Flag(key).Value.Set(val)
				require.NoError(t, err)
				defer func(key string) {
					err := addCmd.Flag(key).Value.Set(oldVal)
					require.NoError(t, err)
				}(key)
			}

			var buffer bytes.Buffer
			color.Output = &buffer

			addCmd.Run(&cobra.Command{}, []string{})
			require.Equal(t, tt.expectedOutput, buffer.String())
		})
	}
}
