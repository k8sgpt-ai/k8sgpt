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

package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// An empty password must not panic when Configure inspects the first byte of
// the token to detect JSON credentials. It should surface a normal error from
// the SDK instead of an index out of range.
func TestGoogleGenAIClientConfigureEmptyPassword(t *testing.T) {
	client := &GoogleGenAIClient{}

	var err error
	require.NotPanics(t, func() {
		err = client.Configure(&AIProvider{Name: googleAIClientName, Password: ""})
	})
	assert.Error(t, err)
}
