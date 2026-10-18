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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/spf13/viper"
)

// Exclusion mutes one known finding by content hash. A muted finding comes
// back automatically when its content changes, so suppressions cannot go
// stale; expires is the escape hatch for findings that are eventually fixed
// upstream (issue #372).
type Exclusion struct {
	Reason  string `mapstructure:"reason"`
	Hash    string `mapstructure:"hash"`
	Expires string `mapstructure:"expires"`
}

// FindingHash derives the exclusion hash of a finding: first 12 hex chars of
// SHA-256 over Kind, the result's namespace/name key, and the whitespace-
// normalized failure text, before any anonymization. Event messages embed
// container-restart counts that drift between runs, so whitespace runs are
// collapsed to keep the hash stable across restarts while still changing when
// the finding's substance changes.
func FindingHash(kind, name, text string) string {
	normalized := strings.Join(strings.Fields(text), " ")
	sum := sha256.Sum256([]byte(kind + "\x00" + name + "\x00" + normalized))
	return hex.EncodeToString(sum[:])[:12]
}

// loadExclusions reads the exclusions key from the viper config, skipping
// expired entries and entries without a reason. Validation failures are
// reported through a.Errors rather than aborting the analysis.
func (a *Analysis) loadExclusions() []Exclusion {
	var exclusions []Exclusion
	if err := viper.UnmarshalKey("exclusions", &exclusions); err != nil {
		a.Errors = append(a.Errors, fmt.Sprintf("[exclusions] %s", err))
		return nil
	}
	now := time.Now()
	var live []Exclusion
	for _, ex := range exclusions {
		if strings.TrimSpace(ex.Reason) == "" {
			a.Errors = append(a.Errors, "[exclusions] an entry without a reason was ignored (a reason is required for every exclusion)")
			continue
		}
		if ex.Expires != "" {
			if expiry, err := time.Parse(time.RFC3339, ex.Expires); err != nil {
				a.Errors = append(a.Errors, fmt.Sprintf("[exclusions] invalid expires %q: %s", ex.Expires, err))
				continue
			} else if !now.Before(expiry) {
				continue // expired: the finding surfaces again
			}
		}
		live = append(live, ex)
	}
	return live
}

// ApplyExclusions drops findings muted by the exclusions config, analyzer-
// agnostic: it runs after RunAnalysis/RunCustomAnalysis and before
// GetAIResults/PrintOutput, so it covers built-in and custom analyzers alike
// and also skips the AI call for muted findings. A Result survives unless
// every failure in it is muted; partially muted results keep their remaining
// failures. The number of dropped failures is reported as a.Excluded.
func (a *Analysis) ApplyExclusions() {
	exclusions := a.loadExclusions()
	if len(exclusions) == 0 {
		return
	}
	hashes := make(map[string]bool, len(exclusions))
	for _, ex := range exclusions {
		hashes[strings.ToLower(ex.Hash)] = true
	}

	kept := make([]common.Result, 0, len(a.Results))
	for _, result := range a.Results {
		keptFailures := result.Error[:0:0]
		for _, failure := range result.Error {
			if hashes[FindingHash(result.Kind, result.Name, failure.Text)] {
				a.Excluded++
			} else {
				keptFailures = append(keptFailures, failure)
			}
		}
		if len(keptFailures) > 0 {
			result.Error = keptFailures
			kept = append(kept, result)
		}
	}
	a.Results = kept
}
