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

package analysis

import (
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func finding(kind, name, text string) common.Result {
	return common.Result{
		Kind: kind,
		Name: name,
		Error: []common.Failure{
			{Text: text},
		},
	}
}

func withExclusions(t *testing.T, entries []map[string]interface{}) {
	t.Helper()
	viper.Set("exclusions", entries)
	t.Cleanup(func() { viper.Set("exclusions", []interface{}{}) })
}

func TestApplyExclusionsDropsMutedFinding(t *testing.T) {
	a := &Analysis{
		Results: []common.Result{
			finding("Pod", "default/web", "the last termination reason is OOMKilled container=app pod=web"),
			finding("Service", "default/api", "Service has no endpoints, expected label app=api"),
		},
	}
	withExclusions(t, []map[string]interface{}{
		{
			"reason": "OOMKilled on restarts is a known downstream effect of the JVM heap config",
			"hash":   FindingHash("Pod", "default/web", "the last termination reason is OOMKilled container=app pod=web"),
		},
	})

	a.ApplyExclusions()

	require.Len(t, a.Results, 1, "only the muted finding is dropped")
	require.Equal(t, "Service", a.Results[0].Kind)
	require.Equal(t, 1, a.Excluded)
	require.Empty(t, a.Errors)
}

func TestApplyExclusionsHashStableAcrossWhitespace(t *testing.T) {
	a := &Analysis{
		Results: []common.Result{
			finding("Pod", "default/web", "the last termination reason  is OOMKilled   container=app pod=web"),
		},
	}
	withExclusions(t, []map[string]interface{}{
		{
			"reason": "known OOMKill",
			"hash":   FindingHash("Pod", "default/web", "the last termination reason is OOMKilled container=app pod=web"),
		},
	})

	a.ApplyExclusions()

	require.Empty(t, a.Results, "whitespace drift between runs must not defeat the mute")
	require.Equal(t, 1, a.Excluded)
}

func TestApplyExclusionsKeepsPartiallyMutedResult(t *testing.T) {
	a := &Analysis{
		Results: []common.Result{
			{
				Kind: "Pod",
				Name: "default/web",
				Error: []common.Failure{
					{Text: "known noise"},
					{Text: "a real problem"},
				},
			},
		},
	}
	withExclusions(t, []map[string]interface{}{
		{"reason": "tracked elsewhere", "hash": FindingHash("Pod", "default/web", "known noise")},
	})

	a.ApplyExclusions()

	require.Len(t, a.Results, 1, "a result with surviving failures is kept")
	require.Len(t, a.Results[0].Error, 1)
	require.Equal(t, "a real problem", a.Results[0].Error[0].Text)
	require.Equal(t, 1, a.Excluded)
}

func TestApplyExclusionsExpiredEntrySurfacesAgain(t *testing.T) {
	a := &Analysis{
		Results: []common.Result{
			finding("Pod", "default/web", "known noise"),
		},
	}
	withExclusions(t, []map[string]interface{}{
		{
			"reason":  "was true last quarter",
			"hash":    FindingHash("Pod", "default/web", "known noise"),
			"expires": "2000-01-01T00:00:00Z",
		},
	})

	a.ApplyExclusions()

	require.Len(t, a.Results, 1, "an expired exclusion must not mute anything")
	require.Equal(t, 0, a.Excluded)
}

func TestApplyExclusionsRejectsEntryWithoutReason(t *testing.T) {
	a := &Analysis{
		Results: []common.Result{
			finding("Pod", "default/web", "known noise"),
		},
	}
	withExclusions(t, []map[string]interface{}{
		{"hash": FindingHash("Pod", "default/web", "known noise")},
	})

	a.ApplyExclusions()

	require.Len(t, a.Results, 1, "a reasonless exclusion is ignored, not applied")
	require.NotEmpty(t, a.Errors, "the config problem is reported to the user")
}
