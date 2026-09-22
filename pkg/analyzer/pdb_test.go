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

package analyzer

import (
	"context"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/stretchr/testify/require"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPodDisruptionBudgetAnalyzer(t *testing.T) {
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: fake.NewSimpleClientset(
				&policyv1.PodDisruptionBudget{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "PDB1",
						Namespace: "test",
					},
					// Status conditions are nil.
					Status: policyv1.PodDisruptionBudgetStatus{
						Conditions: nil,
					},
				},
				&policyv1.PodDisruptionBudget{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "PDB2",
						Namespace: "test",
					},
					// Status conditions are empty.
					Status: policyv1.PodDisruptionBudgetStatus{
						Conditions: []metav1.Condition{},
					},
				},
				&policyv1.PodDisruptionBudget{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "PDB3",
						Namespace: "test",
					},
					Status: policyv1.PodDisruptionBudgetStatus{
						Conditions: []metav1.Condition{
							{
								Type:   "DisruptionAllowed",
								Status: "False",
								Reason: "test reason",
							},
						},
					},
					Spec: policyv1.PodDisruptionBudgetSpec{
						MaxUnavailable: &intstr.IntOrString{
							Type:   0,
							IntVal: 17,
							StrVal: "17",
						},
						MinAvailable: &intstr.IntOrString{
							Type:   0,
							IntVal: 7,
							StrVal: "7",
						},
						// MatchLabels specified.
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"label1": "test1",
								"label2": "test2",
							},
						},
					},
				},
				&policyv1.PodDisruptionBudget{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "PDB4",
						Namespace: "test",
					},
					Status: policyv1.PodDisruptionBudgetStatus{
						Conditions: []metav1.Condition{
							{
								Type:   "DisruptionAllowed",
								Status: "False",
								Reason: "test reason",
							},
						},
					},
					// An empty selector selects every pod in the namespace, so a
					// blocked PDB using one is reported like any other.
					Spec: policyv1.PodDisruptionBudgetSpec{
						Selector: &metav1.LabelSelector{},
					},
				},
			),
		},
		Context:   context.Background(),
		Namespace: "test",
	}

	pdbAnalyzer := PdbAnalyzer{}
	results, err := pdbAnalyzer.Analyze(config)
	require.NoError(t, err)
	names := make([]string, 0, len(results))
	for _, r := range results {
		names = append(names, r.Name)
	}
	// Results are collected from a map, so assert on the set rather than order.
	require.ElementsMatch(t, []string{"test/PDB3", "test/PDB4"}, names)
}

func TestPodDisruptionBudgetAnalyzerLabelSelectorFiltering(t *testing.T) {
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: fake.NewSimpleClientset(
				&policyv1.PodDisruptionBudget{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "PDB1",
						Namespace: "default",
						Labels: map[string]string{
							"app": "pdb",
						},
					},
					// Status conditions are nil.
					Status: policyv1.PodDisruptionBudgetStatus{
						Conditions: []metav1.Condition{
							{
								Type:   "DisruptionAllowed",
								Status: "False",
								Reason: "test reason",
							},
						},
					},
					Spec: policyv1.PodDisruptionBudgetSpec{
						MaxUnavailable: &intstr.IntOrString{
							Type:   0,
							IntVal: 17,
							StrVal: "17",
						},
						MinAvailable: &intstr.IntOrString{
							Type:   0,
							IntVal: 7,
							StrVal: "7",
						},
						// MatchLabels specified.
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"label1": "test1",
								"label2": "test2",
							},
						},
					},
				},
				&policyv1.PodDisruptionBudget{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "PDB2",
						Namespace: "default",
					},
					// Status conditions are empty.
					Status: policyv1.PodDisruptionBudgetStatus{
						Conditions: []metav1.Condition{
							{
								Type:   "DisruptionAllowed",
								Status: "False",
								Reason: "test reason",
							},
						},
					},
					Spec: policyv1.PodDisruptionBudgetSpec{
						MaxUnavailable: &intstr.IntOrString{
							Type:   0,
							IntVal: 17,
							StrVal: "17",
						},
						MinAvailable: &intstr.IntOrString{
							Type:   0,
							IntVal: 7,
							StrVal: "7",
						},
						// MatchLabels specified.
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"label1": "test1",
								"label2": "test2",
							},
						},
					},
				},
			),
		},
		Context:       context.Background(),
		Namespace:     "default",
		LabelSelector: "app=pdb",
	}

	pdbAnalyzer := PdbAnalyzer{}
	results, err := pdbAnalyzer.Analyze(config)
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	require.Equal(t, "default/PDB1", results[0].Name)
}

func TestPodDisruptionBudgetAnalyzerNonMatchLabelsSelectors(t *testing.T) {
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: fake.NewSimpleClientset(
				&policyv1.PodDisruptionBudget{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "empty-selector",
						Namespace: "test",
					},
					Status: policyv1.PodDisruptionBudgetStatus{
						Conditions: []metav1.Condition{
							{
								Type:   "DisruptionAllowed",
								Status: "False",
								Reason: "InsufficientPods",
							},
						},
					},
					Spec: policyv1.PodDisruptionBudgetSpec{
						MinAvailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
						// An empty selector selects every pod in the namespace,
						// so a blocked PDB here has the widest blast radius.
						Selector: &metav1.LabelSelector{},
					},
				},
				&policyv1.PodDisruptionBudget{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "match-expressions",
						Namespace: "test",
					},
					Status: policyv1.PodDisruptionBudgetStatus{
						Conditions: []metav1.Condition{
							{
								Type:   "DisruptionAllowed",
								Status: "False",
								Reason: "InsufficientPods",
							},
						},
					},
					Spec: policyv1.PodDisruptionBudgetSpec{
						MinAvailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
						Selector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "app",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{"web"},
								},
							},
						},
					},
				},
				&policyv1.PodDisruptionBudget{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nil-selector",
						Namespace: "test",
					},
					Status: policyv1.PodDisruptionBudgetStatus{
						Conditions: []metav1.Condition{
							{
								Type:   "DisruptionAllowed",
								Status: "False",
								Reason: "InsufficientPods",
							},
						},
					},
					Spec: policyv1.PodDisruptionBudgetSpec{
						MinAvailable: &intstr.IntOrString{Type: intstr.Int, IntVal: 1},
						// A nil selector matches no pods, so it cannot block
						// an eviction and must not be reported.
						Selector: nil,
					},
				},
			),
		},
		Context:   context.Background(),
		Namespace: "test",
	}

	pdbAnalyzer := PdbAnalyzer{}
	results, err := pdbAnalyzer.Analyze(config)
	require.NoError(t, err)

	names := make([]string, 0, len(results))
	for _, r := range results {
		names = append(names, r.Name)
	}
	require.ElementsMatch(t, []string{"test/empty-selector", "test/match-expressions"}, names)
}
