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
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ValidatingAdmissionPolicyAnalyzer struct{}

func (ValidatingAdmissionPolicyAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	const (
		policyKind  = "ValidatingAdmissionPolicy"
		bindingKind = "ValidatingAdmissionPolicyBinding"
	)

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": policyKind,
	})
	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": bindingKind,
	})

	policyDoc := kubernetes.K8sApiReference{
		Kind: policyKind,
		ApiVersion: schema.GroupVersion{
			Group:   "admissionregistration.k8s.io",
			Version: "v1",
		},
		OpenapiSchema: a.OpenapiSchema,
	}
	bindingDoc := policyDoc
	bindingDoc.Kind = bindingKind

	client := a.Client.GetClient().AdmissionregistrationV1()
	listOptions := metav1.ListOptions{LabelSelector: a.LabelSelector}

	policies, err := client.ValidatingAdmissionPolicies().List(a.Context, listOptions)
	if err != nil {
		return nil, err
	}

	var results []common.Result
	for _, policy := range policies.Items {
		if policy.Status.TypeChecking == nil {
			continue
		}

		var failures []common.Failure
		for _, warning := range policy.Status.TypeChecking.ExpressionWarnings {
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf(
					"ValidatingAdmissionPolicy %s has a CEL type-check warning at %s: %s",
					policy.Name,
					warning.FieldRef,
					warning.Warning,
				),
				KubernetesDoc: policyDoc.GetApiDocV2("status.typeChecking.expressionWarnings"),
				Sensitive: []common.Sensitive{
					{
						Unmasked: policy.Name,
						Masked:   util.MaskString(policy.Name),
					},
				},
			})
		}

		if len(failures) == 0 {
			continue
		}

		results = append(results, common.Result{
			Kind:  policyKind,
			Name:  policy.Name,
			Error: failures,
		})
		AnalyzerErrorsMetric.WithLabelValues(policyKind, policy.Name, "").Set(float64(len(failures)))
	}

	bindings, err := client.ValidatingAdmissionPolicyBindings().List(a.Context, listOptions)
	if err != nil {
		return nil, err
	}

	policyExists := make(map[string]bool)
	policyChecked := make(map[string]bool)
	for _, binding := range bindings.Items {
		policyName := binding.Spec.PolicyName
		if !policyChecked[policyName] {
			policyChecked[policyName] = true
			if policyName != "" {
				_, getErr := client.ValidatingAdmissionPolicies().Get(a.Context, policyName, metav1.GetOptions{})
				switch {
				case getErr == nil:
					policyExists[policyName] = true
				case apierrors.IsNotFound(getErr):
					policyExists[policyName] = false
				default:
					return nil, getErr
				}
			}
		}

		if policyExists[policyName] {
			continue
		}

		failure := common.Failure{
			Text: fmt.Sprintf(
				"ValidatingAdmissionPolicyBinding %s references ValidatingAdmissionPolicy %s which does not exist.",
				binding.Name,
				policyName,
			),
			KubernetesDoc: bindingDoc.GetApiDocV2("spec.policyName"),
			Sensitive: []common.Sensitive{
				{
					Unmasked: binding.Name,
					Masked:   util.MaskString(binding.Name),
				},
				{
					Unmasked: policyName,
					Masked:   util.MaskString(policyName),
				},
			},
		}
		results = append(results, common.Result{
			Kind:  bindingKind,
			Name:  binding.Name,
			Error: []common.Failure{failure},
		})
		AnalyzerErrorsMetric.WithLabelValues(bindingKind, binding.Name, "").Set(1)
	}

	return results, nil
}
