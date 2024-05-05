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

	openapi_v2 "github.com/google/gnostic/openapiv2"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestGetApiDocV2(t *testing.T) {
	k8s := &K8sApiReference{
		ApiVersion: schema.GroupVersion{
			Group:   "group.v1",
			Version: "v1",
		},
		OpenapiSchema: &openapi_v2.Document{
			Definitions: &openapi_v2.Definitions{
				AdditionalProperties: []*openapi_v2.NamedSchema{
					{
						Name: "group.v1.kind",
						Value: &openapi_v2.Schema{
							Title: "test",
							Properties: &openapi_v2.Properties{
								AdditionalProperties: []*openapi_v2.NamedSchema{
									{
										Name: "schema1",
										Value: &openapi_v2.Schema{
											Title:       "test",
											Description: "schema1 description",
											Type: &openapi_v2.TypeItem{
												Value: []string{"string"},
											},
										},
									},
									{
										Name: "schema2",
										Value: &openapi_v2.Schema{
											Items: &openapi_v2.ItemsItem{
												Schema: []*openapi_v2.Schema{
													{
														Title: "random-schema",
													},
												},
											},
											Title:       "test",
											XRef:        "xref",
											Description: "schema2 description",
											Type: &openapi_v2.TypeItem{
												Value: []string{"bool"},
											},
										},
									},
								},
							},
						},
					},
					{
						Name: "group",
					},
				},
			},
		},
		Kind: "kind",
	}

	tests := []struct {
		name           string
		field          string
		expectedOutput string
	}{
		{
			name: "empty field",
		},
		{
			name:           "2 schemas",
			field:          "schema2.schema1",
			expectedOutput: "",
		},
		{
			name:           "schema1 description",
			field:          "schema1",
			expectedOutput: "schema1 description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := k8s.GetApiDocV2(tt.field)
			require.Equal(t, tt.expectedOutput, output)
		})
	}
}
