/*
Copyright 2026 The K8sGPT Authors.
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
	"context"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/yaml"
)

const fpCorpusDir = "testdata/fpcorpus"

// fpCase is the header document of a false-positive corpus case.
type fpCase struct {
	Issues        []string `json:"issues"`
	Description   string   `json:"description"`
	Namespace     string   `json:"namespace"`
	ExpectResults []string `json:"expectResults"`
}

var yamlDocSeparator = regexp.MustCompile(`(?m)^---\s*$`)

// loadFPCase reads a corpus file and returns its header and the Kubernetes
// objects that follow it.
func loadFPCase(t *testing.T, path string) (fpCase, []runtime.Object) {
	t.Helper()
	raw, err := os.ReadFile(path)
	require.NoError(t, err)

	docs := yamlDocSeparator.Split(string(raw), -1)
	require.GreaterOrEqual(t, len(docs), 2, "%s: needs a header document and at least one object", path)

	var c fpCase
	require.NoError(t, yaml.UnmarshalStrict([]byte(docs[0]), &c), "%s: invalid case header", path)
	require.NotEmpty(t, c.Description, "%s: description is required", path)

	decode := scheme.Codecs.UniversalDeserializer().Decode
	var objs []runtime.Object
	for i, doc := range docs[1:] {
		if strings.TrimSpace(doc) == "" {
			continue
		}
		obj, _, err := decode([]byte(doc), nil, nil)
		require.NoError(t, err, "%s: object %d", path, i+1)
		objs = append(objs, obj)
	}
	return c, objs
}

// TestFalsePositiveCorpus runs every case under testdata/fpcorpus against the
// analyzer named by its directory and checks the reported results.
func TestFalsePositiveCorpus(t *testing.T) {
	_, analyzers := GetAnalyzerMap()

	files, err := filepath.Glob(filepath.Join(fpCorpusDir, "*", "*.yaml"))
	require.NoError(t, err)
	require.NotEmpty(t, files, "no corpus cases found")

	for _, file := range files {
		analyzerName := filepath.Base(filepath.Dir(file))
		name := analyzerName + "/" + strings.TrimSuffix(filepath.Base(file), ".yaml")

		t.Run(name, func(t *testing.T) {
			a, ok := analyzers[analyzerName]
			require.True(t, ok, "unknown analyzer %q (directory name must match a registered analyzer)", analyzerName)

			c, objs := loadFPCase(t, file)
			results, err := a.Analyze(common.Analyzer{
				Client:    &kubernetes.Client{Client: fake.NewSimpleClientset(objs...)},
				Context:   context.Background(),
				Namespace: c.Namespace,
			})
			require.NoError(t, err)

			got := make([]string, 0, len(results))
			for _, r := range results {
				got = append(got, r.Name)
			}
			sort.Strings(got)
			want := append([]string{}, c.ExpectResults...)
			sort.Strings(want)

			require.Equal(t, want, got, "%s (%s): %s\nresults: %+v",
				name, strings.Join(c.Issues, ", "), c.Description, results)
		})
	}
}
