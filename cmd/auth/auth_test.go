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
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestAuthCmd(t *testing.T) {
	require.Equal(t, "auth", AuthCmd.Name())

	// Redirect stdout to a buffer
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	want := "This is a test command help function\n"
	cmd := &cobra.Command{}
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Print(want)
	})
	AuthCmd.Run(cmd, []string{})

	w.Close()
	got, _ := io.ReadAll(r)
	os.Stdout = rescueStdout

	require.Equal(t, want, string(got))
}
