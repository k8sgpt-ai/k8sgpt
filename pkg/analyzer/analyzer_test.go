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

package analyzer

import (
	"sort"
	"testing"
)

// keysOf returns the sorted keys of an analyzer map, so failures read as a
// diffable list rather than as map iteration order.
func keysOf[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

// ListFilters drives `k8sgpt filters list` and the values --filter accepts, so
// every registered analyzer must be advertised by it.
func TestListFilters_ReportsEveryRegisteredAnalyzer(t *testing.T) {
	coreKeys, additionalKeys, _ := ListFilters()

	for _, want := range keysOf(coreAnalyzerMap) {
		if !contains(coreKeys, want) {
			t.Errorf("core analyzer %q is registered but not returned by ListFilters", want)
		}
	}
	for _, want := range keysOf(additionalAnalyzerMap) {
		if !contains(additionalKeys, want) {
			t.Errorf("additional analyzer %q is registered but not returned by ListFilters", want)
		}
	}

	if len(coreKeys) != len(coreAnalyzerMap) {
		t.Errorf("ListFilters returned %d core filters, want %d", len(coreKeys), len(coreAnalyzerMap))
	}
	if len(additionalKeys) != len(additionalAnalyzerMap) {
		t.Errorf("ListFilters returned %d additional filters, want %d",
			len(additionalKeys), len(additionalAnalyzerMap))
	}
}

// A name in both maps would make the merged map ambiguous: the additional entry
// silently wins in GetAnalyzerMap while ListFilters advertises the name twice.
func TestAnalyzerMaps_DoNotOverlap(t *testing.T) {
	for name := range additionalAnalyzerMap {
		if _, clash := coreAnalyzerMap[name]; clash {
			t.Errorf("analyzer %q is registered as both core and additional; "+
				"the additional entry would shadow the core one in GetAnalyzerMap", name)
		}
	}
}

// No entry may be nil: a nil analyzer passes registration and only fails later,
// at analysis time, as a panic.
func TestAnalyzerMaps_HaveNoNilEntries(t *testing.T) {
	for name, a := range coreAnalyzerMap {
		if a == nil {
			t.Errorf("core analyzer %q is registered as nil", name)
		}
	}
	for name, a := range additionalAnalyzerMap {
		if a == nil {
			t.Errorf("additional analyzer %q is registered as nil", name)
		}
	}
}

func TestGetAnalyzerMap_CoreMapMatchesRegistration(t *testing.T) {
	core, _ := GetAnalyzerMap()

	if got, want := len(core), len(coreAnalyzerMap); got != want {
		t.Errorf("core map has %d analyzers, want %d", got, want)
	}
	for name := range coreAnalyzerMap {
		if _, ok := core[name]; !ok {
			t.Errorf("core analyzer %q missing from GetAnalyzerMap's core map", name)
		}
	}
}

// The merged map must be a superset of both registries; anything missing here
// is an analyzer that can never run.
func TestGetAnalyzerMap_MergedContainsCoreAndAdditional(t *testing.T) {
	core, merged := GetAnalyzerMap()

	for name := range coreAnalyzerMap {
		if _, ok := merged[name]; !ok {
			t.Errorf("core analyzer %q missing from the merged map", name)
		}
	}
	for name := range additionalAnalyzerMap {
		if _, ok := merged[name]; !ok {
			t.Errorf("additional analyzer %q missing from the merged map", name)
		}
	}
	if len(merged) < len(core) {
		t.Errorf("merged map (%d) is smaller than the core map (%d)", len(merged), len(core))
	}
}

// Mutating a returned map must not corrupt the package-level registries, or one
// caller could disable an analyzer for every later caller in the same process.
func TestGetAnalyzerMap_ReturnsIndependentCopies(t *testing.T) {
	coreBefore := len(coreAnalyzerMap)

	core, merged := GetAnalyzerMap()
	for k := range core {
		delete(core, k)
	}
	for k := range merged {
		delete(merged, k)
	}

	if got := len(coreAnalyzerMap); got != coreBefore {
		t.Errorf("coreAnalyzerMap was mutated through the returned map: %d entries, want %d",
			got, coreBefore)
	}
	if freshCore, _ := GetAnalyzerMap(); len(freshCore) != coreBefore {
		t.Errorf("a later GetAnalyzerMap call returned %d core analyzers, want %d",
			len(freshCore), coreBefore)
	}
}

// The filters a user can pass must be exactly the analyzers that can run.
// Registering an analyzer in one place and forgetting the other yields either a
// filter that advertises an analyzer that never runs, or an analyzer nobody can
// select.
func TestListFiltersAndGetAnalyzerMap_AgreeOnCoreAndAdditional(t *testing.T) {
	coreKeys, additionalKeys, _ := ListFilters()
	_, merged := GetAnalyzerMap()

	for _, name := range append(append([]string{}, coreKeys...), additionalKeys...) {
		if _, ok := merged[name]; !ok {
			t.Errorf("ListFilters advertises %q but GetAnalyzerMap cannot run it", name)
		}
	}
	for _, name := range keysOf(coreAnalyzerMap) {
		if !contains(coreKeys, name) {
			t.Errorf("analyzer %q can run but is not advertised as a core filter", name)
		}
	}
	for _, name := range keysOf(additionalAnalyzerMap) {
		if !contains(additionalKeys, name) {
			t.Errorf("analyzer %q can run but is not advertised as an additional filter", name)
		}
	}
}
