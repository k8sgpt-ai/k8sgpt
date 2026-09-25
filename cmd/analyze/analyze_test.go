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

package analyze

import (
	"strings"
	"testing"
)

// --resource hands the name to every analyzer that runs, so the filter set and
// the resource kind must agree; a mismatched --filter would make unrelated
// analyzers select metadata.name for a kind the user never asked about.
func TestResolveResourceSelection(t *testing.T) {
	tests := []struct {
		name         string
		resource     string
		filters      []string
		wantFilters  []string
		wantResource string
		wantErr      string
	}{
		{
			name:        "no resource flag leaves filters untouched",
			filters:     []string{"Deployment", "Service"},
			wantFilters: []string{"Deployment", "Service"},
		},
		{
			name:         "resource alone selects its kind",
			resource:     "Deployment/my-app",
			wantFilters:  []string{"Deployment"},
			wantResource: "my-app",
		},
		{
			name:         "matching filter reduces to the resource kind",
			resource:     "Deployment/my-app",
			filters:      []string{"Deployment"},
			wantFilters:  []string{"Deployment"},
			wantResource: "my-app",
		},
		{
			name:         "matching filter is case-insensitive",
			resource:     "Deployment/my-app",
			filters:      []string{"deployment"},
			wantFilters:  []string{"Deployment"},
			wantResource: "my-app",
		},
		{
			name:     "filter for a different kind is rejected",
			resource: "Deployment/my-app",
			filters:  []string{"Service"},
			wantErr:  `also requests "Service"`,
		},
		{
			name:     "multi-kind filter set is rejected, not silently narrowed",
			resource: "Deployment/my-app",
			filters:  []string{"Deployment", "Service"},
			wantErr:  `also requests "Service"`,
		},
		{
			name:     "missing name half is rejected",
			resource: "Deployment",
			wantErr:  "Kind/Name format",
		},
		{
			name:     "missing kind half is rejected",
			resource: "/my-app",
			wantErr:  "Kind/Name format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFilters, gotResource, err := resolveResourceSelection(tt.resource, tt.filters)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotResource != tt.wantResource {
				t.Errorf("resource name = %q, want %q", gotResource, tt.wantResource)
			}
			if len(gotFilters) != len(tt.wantFilters) {
				t.Fatalf("filters = %v, want %v", gotFilters, tt.wantFilters)
			}
			for i := range gotFilters {
				if gotFilters[i] != tt.wantFilters[i] {
					t.Errorf("filters = %v, want %v", gotFilters, tt.wantFilters)
				}
			}
		})
	}
}
