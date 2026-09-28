package common

import (
	"testing"

	"k8s.io/apimachinery/pkg/fields"
)

// The --resource flag must push the name predicate down to the API server as a
// field selector, and must not change list behaviour when it is unset.
func TestListOptions(t *testing.T) {
	tests := []struct {
		name              string
		labelSelector     string
		resourceName      string
		wantLabelSelector string
		wantFieldSelector string
	}{
		{
			name:              "absent resource name preserves existing behaviour",
			labelSelector:     "app=demo",
			resourceName:      "",
			wantLabelSelector: "app=demo",
			wantFieldSelector: "",
		},
		{
			name:              "no selectors at all",
			wantLabelSelector: "",
			wantFieldSelector: "",
		},
		{
			name:              "resource name becomes a metadata.name field selector",
			resourceName:      "alpha-api",
			wantLabelSelector: "",
			wantFieldSelector: fields.OneTermEqualSelector("metadata.name", "alpha-api").String(),
		},
		{
			name:              "label selector and resource name combine",
			labelSelector:     "tier=web",
			resourceName:      "beta-worker",
			wantLabelSelector: "tier=web",
			wantFieldSelector: fields.OneTermEqualSelector("metadata.name", "beta-worker").String(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Analyzer{LabelSelector: tt.labelSelector, ResourceName: tt.resourceName}
			got := a.ListOptions()
			if got.LabelSelector != tt.wantLabelSelector {
				t.Errorf("LabelSelector = %q, want %q", got.LabelSelector, tt.wantLabelSelector)
			}
			if got.FieldSelector != tt.wantFieldSelector {
				t.Errorf("FieldSelector = %q, want %q", got.FieldSelector, tt.wantFieldSelector)
			}
		})
	}
}
