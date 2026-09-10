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
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/stretchr/testify/require"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestValidatingAdmissionPolicyAnalyzer(t *testing.T) {
	config := newValidatingAdmissionPolicyAnalyzerConfig(
		&admissionregistrationv1.ValidatingAdmissionPolicy{
			ObjectMeta: metav1.ObjectMeta{Name: "valid-policy"},
		},
		&admissionregistrationv1.ValidatingAdmissionPolicy{
			ObjectMeta: metav1.ObjectMeta{Name: "warning-policy"},
			Status: admissionregistrationv1.ValidatingAdmissionPolicyStatus{
				TypeChecking: &admissionregistrationv1.TypeChecking{
					ExpressionWarnings: []admissionregistrationv1.ExpressionWarning{
						{
							FieldRef: "spec.validations[0].expression",
							Warning:  "no matching overload for equals",
						},
						{
							FieldRef: "spec.variables[0].expression",
							Warning:  "undefined field 'team'",
						},
					},
				},
			},
		},
		&admissionregistrationv1.ValidatingAdmissionPolicyBinding{
			ObjectMeta: metav1.ObjectMeta{Name: "valid-binding"},
			Spec: admissionregistrationv1.ValidatingAdmissionPolicyBindingSpec{
				PolicyName: "valid-policy",
			},
		},
		&admissionregistrationv1.ValidatingAdmissionPolicyBinding{
			ObjectMeta: metav1.ObjectMeta{Name: "dangling-binding"},
			Spec: admissionregistrationv1.ValidatingAdmissionPolicyBindingSpec{
				PolicyName: "missing-policy",
			},
		},
	)

	results, err := (ValidatingAdmissionPolicyAnalyzer{}).Analyze(config)
	require.NoError(t, err)
	require.Len(t, results, 2)

	policyResult := requireAnalyzerResult(t, results, "ValidatingAdmissionPolicy", "warning-policy")
	require.Len(t, policyResult.Error, 2)
	require.Equal(
		t,
		"ValidatingAdmissionPolicy warning-policy has a CEL type-check warning at "+
			"spec.validations[0].expression: no matching overload for equals",
		policyResult.Error[0].Text,
	)
	require.Equal(
		t,
		"ValidatingAdmissionPolicy warning-policy has a CEL type-check warning at "+
			"spec.variables[0].expression: undefined field 'team'",
		policyResult.Error[1].Text,
	)

	bindingResult := requireAnalyzerResult(t, results, "ValidatingAdmissionPolicyBinding", "dangling-binding")
	require.Len(t, bindingResult.Error, 1)
	require.Equal(
		t,
		"ValidatingAdmissionPolicyBinding dangling-binding references "+
			"ValidatingAdmissionPolicy missing-policy which does not exist.",
		bindingResult.Error[0].Text,
	)
}

func TestValidatingAdmissionPolicyAnalyzerLabelSelector(t *testing.T) {
	config := newValidatingAdmissionPolicyAnalyzerConfig(
		&admissionregistrationv1.ValidatingAdmissionPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "selected-warning-policy",
				Labels: map[string]string{"analyze": "true"},
			},
			Status: admissionregistrationv1.ValidatingAdmissionPolicyStatus{
				TypeChecking: &admissionregistrationv1.TypeChecking{
					ExpressionWarnings: []admissionregistrationv1.ExpressionWarning{
						{
							FieldRef: "spec.validations[0].expression",
							Warning:  "selected warning",
						},
					},
				},
			},
		},
		&admissionregistrationv1.ValidatingAdmissionPolicy{
			ObjectMeta: metav1.ObjectMeta{Name: "unselected-policy"},
			Status: admissionregistrationv1.ValidatingAdmissionPolicyStatus{
				TypeChecking: &admissionregistrationv1.TypeChecking{
					ExpressionWarnings: []admissionregistrationv1.ExpressionWarning{
						{
							FieldRef: "spec.validations[0].expression",
							Warning:  "unselected warning",
						},
					},
				},
			},
		},
		&admissionregistrationv1.ValidatingAdmissionPolicyBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "selected-valid-binding",
				Labels: map[string]string{"analyze": "true"},
			},
			Spec: admissionregistrationv1.ValidatingAdmissionPolicyBindingSpec{
				PolicyName: "unselected-policy",
			},
		},
		&admissionregistrationv1.ValidatingAdmissionPolicyBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "selected-dangling-binding",
				Labels: map[string]string{"analyze": "true"},
			},
			Spec: admissionregistrationv1.ValidatingAdmissionPolicyBindingSpec{
				PolicyName: "missing-policy",
			},
		},
		&admissionregistrationv1.ValidatingAdmissionPolicyBinding{
			ObjectMeta: metav1.ObjectMeta{Name: "unselected-dangling-binding"},
			Spec: admissionregistrationv1.ValidatingAdmissionPolicyBindingSpec{
				PolicyName: "another-missing-policy",
			},
		},
	)
	config.LabelSelector = "analyze=true"

	results, err := (ValidatingAdmissionPolicyAnalyzer{}).Analyze(config)
	require.NoError(t, err)
	require.Len(t, results, 2)
	requireAnalyzerResult(t, results, "ValidatingAdmissionPolicy", "selected-warning-policy")
	requireAnalyzerResult(t, results, "ValidatingAdmissionPolicyBinding", "selected-dangling-binding")
}

func TestValidatingAdmissionPolicyAnalyzerIsOptional(t *testing.T) {
	_, isCore := coreAnalyzerMap["ValidatingAdmissionPolicy"]
	require.False(t, isCore)

	registered, isAdditional := additionalAnalyzerMap["ValidatingAdmissionPolicy"]
	require.True(t, isAdditional)
	require.IsType(t, ValidatingAdmissionPolicyAnalyzer{}, registered)
}

func newValidatingAdmissionPolicyAnalyzerConfig(objects ...runtime.Object) common.Analyzer {
	return common.Analyzer{
		Client: &kubernetes.Client{
			Client: fake.NewSimpleClientset(objects...),
		},
		Context: context.Background(),
	}
}

func requireAnalyzerResult(t *testing.T, results []common.Result, kind, name string) common.Result {
	t.Helper()

	for _, result := range results {
		if result.Kind == kind && result.Name == name {
			return result
		}
	}

	require.Failf(t, "analyzer result not found", "kind=%s name=%s", kind, name)
	return common.Result{}
}
