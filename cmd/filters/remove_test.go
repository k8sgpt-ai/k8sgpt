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
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestRemoveCmd(t *testing.T) {
	require.Equal(t, "remove", removeCmd.Name())

	err := removeCmd.Args(&cobra.Command{}, []string{"arg1"})
	require.NoError(t, err)

	err = removeCmd.Args(&cobra.Command{}, []string{"arg1", "arg2"})
	require.ErrorContains(t, err, "accepts 1 arg(s), received 2")

	// Set the configuration file in viper
	configFileName := "delete-config.json"
	data := map[string]interface{}{
		"active_filters": []string{
			"Service",
			"Deployment",
			"Ingress",
			"CronJob",
			"MutatingWebhookConfiguration",
			"Node",
			"ValidatingWebhookConfiguration",
			"PersistentVolumeClaim",
			"StatefulSet",
			"Pod",
			"ReplicaSet",
		},
	}
	err = createConfigFile(data, configFileName)
	require.NoError(t, err)
	defer os.Remove(configFileName)

	viper.SetConfigType("json")
	viper.SetConfigFile(configFileName)
	err = viper.ReadInConfig()
	require.NoError(t, err)

	// Redirect the output of the color functions to buffer.
	var buffer bytes.Buffer
	color.Output = &buffer

	// Remove a filter
	removeCmd.Run(&cobra.Command{}, []string{"MutatingWebhookConfiguration"})
	want := "Filter(s) MutatingWebhookConfiguration removed\n"
	require.Equal(t, want, buffer.String())

	// Initially the list contained 12 filters, after deleting one filter in this
	// test, now it should be left with 11 filters only.
	require.Equal(t, 11, len(viper.GetStringSlice("active_filters")))
}
