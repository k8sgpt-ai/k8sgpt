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

package filters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestAddCmd(t *testing.T) {
	require.Equal(t, "add", addCmd.Name())

	// Set the configuration file in viper
	configFileName := "config.json"
	err := createConfigFile(map[string]interface{}{}, configFileName)
	require.NoError(t, err)
	defer os.Remove(configFileName)

	// Set the configuration file.
	viper.SetConfigType("json")
	viper.SetConfigFile(configFileName)
	err = viper.ReadInConfig()
	require.NoError(t, err)

	// Redirect the output of the color functions to buffer.
	var buffer bytes.Buffer
	color.Output = &buffer

	// Add a filter
	addCmd.Run(&cobra.Command{}, []string{"HorizontalPodAutoScaler"})
	want := "Filter HorizontalPodAutoScaler added\n"
	require.Equal(t, want, buffer.String())

	// Check the length of the active filter, it should be length of core
	// filters + 1 that was added in this test.
	coreFilters, _, _ := analyzer.ListFilters()
	require.Equal(t, len(coreFilters)+1, len(viper.GetStringSlice("active_filters")))
}

func createConfigFile(data map[string]interface{}, configName string) error {
	// Create a config file
	file, err := os.Create(configName)
	if err != nil {
		return err
	}
	// Ensure file is closed
	defer file.Close()

	// Marshal the data into a JSON byte slice
	jsonData, err := json.MarshalIndent(data, "", "  ") // Indent for readability
	if err != nil {
		return fmt.Errorf("error marshalling data: %w", err)
	}

	// Write the JSON data to the file
	_, err = file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("error writing data to file: %w", err)
	}

	return nil
}
