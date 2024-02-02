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

package analysis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrintOutput(t *testing.T) {
	require.NotEmpty(t, getOutputFormats())

	tests := []struct {
		name           string
		a              *Analysis
		format         string
		expectedOutput string
		expectedErr    string
	}{
		{
			name:           "json format",
			a:              &Analysis{},
			format:         "json",
			expectedOutput: "{\n  \"provider\": \"\",\n  \"errors\": null,\n  \"status\": \"OK\",\n  \"problems\": 0,\n  \"results\": null\n}",
		},
		{
			name:           "text format",
			a:              &Analysis{},
			format:         "text",
			expectedOutput: "AI Provider: AI not used; --explain not set\n\nNo problems detected\n",
		},
		{
			name:        "unsupported format",
			a:           &Analysis{},
			format:      "unsupported",
			expectedErr: "unsupported output format: unsupported. Available format",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			output, err := tt.a.PrintOutput(tt.format)
			if tt.expectedErr == "" {
				require.NoError(t, err)
				require.Contains(t, string(output), tt.expectedOutput)
			} else {
				require.ErrorContains(t, err, tt.expectedErr)
				require.Nil(t, output)
			}
		})
	}
}
