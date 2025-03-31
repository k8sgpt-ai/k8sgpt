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

package kyverno

import (
	"context"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/crd/api/policyreport/v1alpha2"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func buildFakeClient(t *testing.T) client.Client {
	objects := []client.Object{
		&v1alpha2.PolicyReport{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "policy-1",
				Namespace: "test-ns",
			},
			Results: []v1alpha2.PolicyReportResult{
				{
					Category: "Other",
					Message:  "validation failure: Images built more than 6 months ago are prohibited.",
					Policy:   "block-stale-images",
					Result:   "fail",
				},
			},
		},
		&v1alpha2.PolicyReport{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "policy-2",
				Namespace: "other-ns",
			},
			Results: []v1alpha2.PolicyReportResult{
				{
					Category: "Other",
					Message:  "validation failure: Images built more than 6 months ago are prohibited.",
					Policy:   "block-stale-images",
					Result:   "fail",
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	err := v1alpha2.AddToScheme(scheme)
	if err != nil {
		t.Error(err)
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

func TestAnalyzerNamespaceFiltering(t *testing.T) {

	config := common.Analyzer{
		Client: &kubernetes.Client{
			CtrlClient: buildFakeClient(t),
		},
		Context:   context.Background(),
		Namespace: "test-ns",
	}

	// Create and run analyzer
	analyzer := KyvernoAnalyzer{
		policyReportAnalysis: true,
	}
	results, err := analyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}

	// Verify results
	assert.Equal(t, len(results), 1)
	assert.Equal(t, results[0].Kind, "PolicyReport")
	assert.Equal(t, results[0].Name, "test-ns/policy-1")
}

func TestAnalyzerAllNamespace(t *testing.T) {

	config := common.Analyzer{
		Client: &kubernetes.Client{
			CtrlClient: buildFakeClient(t),
		},
		Context: context.Background(),
	}

	// Create and run analyzer
	analyzer := KyvernoAnalyzer{
		policyReportAnalysis: true,
	}
	results, err := analyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}

	// Verify results
	assert.Equal(t, len(results), 2)

}
