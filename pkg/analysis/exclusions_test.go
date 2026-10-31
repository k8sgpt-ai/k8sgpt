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
	"encoding/json"
	"testing"
	"time"

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

func TestApplyExclusionsAcceptsDateOnlyExpiry(t *testing.T) {
	// The PR description and the design note both document expires as a plain
	// date, so a date-only value must be accepted instead of being dropped as a
	// config error.
	future := time.Now().AddDate(1, 0, 0).Format(time.DateOnly)
	a := &Analysis{
		Results: []common.Result{
			finding("Pod", "default/web", "known noise"),
		},
	}
	withExclusions(t, []map[string]interface{}{
		{
			"reason":  "tracked upstream, expires next year",
			"hash":    FindingHash("Pod", "default/web", "known noise"),
			"expires": future,
		},
	})

	a.ApplyExclusions()

	require.Empty(t, a.Errors, "a documented date-only value is not a config error")
	require.Empty(t, a.Results, "an unexpired date-only exclusion still mutes")
	require.Equal(t, 1, a.Excluded)
}

func TestApplyExclusionsDateOnlyExpiryCoversTheWholeDay(t *testing.T) {
	// A date-only expiry means "live until this day ends", so an exclusion dated
	// today must keep muting for the rest of today.
	today := time.Now().Format(time.DateOnly)
	a := &Analysis{
		Results: []common.Result{
			finding("Pod", "default/web", "known noise"),
		},
	}
	withExclusions(t, []map[string]interface{}{
		{
			"reason":  "muted until the end of today",
			"hash":    FindingHash("Pod", "default/web", "known noise"),
			"expires": today,
		},
	})

	a.ApplyExclusions()

	require.Empty(t, a.Errors)
	require.Empty(t, a.Results, "an exclusion dated today still covers today")
}

func TestApplyExclusionsRejectsMalformedExpiry(t *testing.T) {
	a := &Analysis{
		Results: []common.Result{
			finding("Pod", "default/web", "known noise"),
		},
	}
	withExclusions(t, []map[string]interface{}{
		{
			"reason":  "bad timestamp",
			"hash":    FindingHash("Pod", "default/web", "known noise"),
			"expires": "next tuesday",
		},
	})

	a.ApplyExclusions()

	require.Len(t, a.Results, 1, "a malformed expiry must not mute")
	require.NotEmpty(t, a.Errors, "the config problem is reported to the user")
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

func TestTextOutputPrintsFindingHash(t *testing.T) {
	// Users need the hash to write an exclusion without recomputing SHA-256 by
	// hand, so text output has to carry it on the finding line.
	const text = "the last termination reason is OOMKilled container=app pod=web"
	a := &Analysis{
		Results: []common.Result{
			finding("Pod", "default/web", text),
		},
	}

	out, err := a.PrintOutput("text")
	require.NoError(t, err)
	require.Contains(t, string(out), "hash="+FindingHash("Pod", "default/web", text),
		"the printed hash is the value the exclusions config matches on")
}

func TestTextOutputReportsExcludedCount(t *testing.T) {
	a := &Analysis{
		Results: []common.Result{
			finding("Service", "default/api", "Service has no endpoints, expected label app=api"),
		},
		Excluded: 2,
	}

	out, err := a.PrintOutput("text")
	require.NoError(t, err)
	require.Contains(t, string(out), "2 findings excluded",
		"the mute count is reported in text output, not only in JSON")
}

func TestJsonOutputCarriesFindingHash(t *testing.T) {
	const text = "known noise"
	a := &Analysis{
		Results: []common.Result{
			finding("Pod", "default/web", text),
		},
	}

	out, err := a.PrintOutput("json")
	require.NoError(t, err)

	var decoded JsonOutput
	require.NoError(t, json.Unmarshal(out, &decoded))
	require.Len(t, decoded.Results, 1)
	require.Len(t, decoded.Results[0].Error, 1)
	require.Equal(t, FindingHash("Pod", "default/web", text), decoded.Results[0].Error[0].Hash,
		"json output exposes the hash as its own field")
}

func TestOutputDoesNotMutateAnalysisResults(t *testing.T) {
	// The hash is a rendering concern: keeping it out of a.Results means it can
	// never leak into the AI provider payload or the exclusion matching itself.
	const text = "known noise"
	a := &Analysis{
		Results: []common.Result{
			finding("Pod", "default/web", text),
		},
	}

	_, err := a.PrintOutput("text")
	require.NoError(t, err)
	_, err = a.PrintOutput("json")
	require.NoError(t, err)

	require.Empty(t, a.Results[0].Error[0].Hash, "rendering must not mutate the analysis")
}
