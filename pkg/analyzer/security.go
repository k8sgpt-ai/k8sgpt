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
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SecurityAnalyzer struct{}

func (SecurityAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	kind := "Security"

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	var results []common.Result

	// Analyze ServiceAccounts
	saResults, err := analyzeServiceAccounts(a)
	if err != nil {
		return nil, err
	}
	results = append(results, saResults...)

	// Analyze RoleBindings
	rbResults, err := analyzeRoleBindings(a)
	if err != nil {
		return nil, err
	}
	results = append(results, rbResults...)

	// Analyze Pod Security Contexts
	podResults, err := analyzePodSecurityContexts(a)
	if err != nil {
		return nil, err
	}
	results = append(results, podResults...)

	return results, nil
}

func analyzeServiceAccounts(a common.Analyzer) ([]common.Result, error) {
	var results []common.Result

	sas, err := a.Client.GetClient().CoreV1().ServiceAccounts(a.Namespace).List(a.Context, metav1.ListOptions{
		LabelSelector: a.LabelSelector,
	})
	if err != nil {
		return nil, err
	}

	for _, sa := range sas.Items {
		var failures []common.Failure

		// Check for default service account usage
		if sa.Name == "default" {
			pods, err := a.Client.GetClient().CoreV1().Pods(sa.Namespace).List(a.Context, metav1.ListOptions{})
			if err != nil {
				continue
			}

			defaultSAUsers := []string{}
			for _, pod := range pods.Items {
				if pod.Spec.ServiceAccountName == "default" {
					defaultSAUsers = append(defaultSAUsers, pod.Name)
				}
			}

			if len(defaultSAUsers) > 0 {
				failures = append(failures, common.Failure{
					Text:      fmt.Sprintf("Default service account is being used by pods: %v", defaultSAUsers),
					Sensitive: []common.Sensitive{},
				})
			}
		}

		if len(failures) > 0 {
			results = append(results, common.Result{
				Kind:  "Security/ServiceAccount",
				Name:  fmt.Sprintf("%s/%s", sa.Namespace, sa.Name),
				Error: failures,
			})
			AnalyzerErrorsMetric.WithLabelValues("Security/ServiceAccount", sa.Name, sa.Namespace).Set(float64(len(failures)))
		}
	}

	return results, nil
}

func analyzeRoleBindings(a common.Analyzer) ([]common.Result, error) {
	var results []common.Result

	rbs, err := a.Client.GetClient().RbacV1().RoleBindings(a.Namespace).List(a.Context, metav1.ListOptions{
		LabelSelector: a.LabelSelector,
	})
	if err != nil {
		return nil, err
	}

	for _, rb := range rbs.Items {
		var failures []common.Failure

		// Check for wildcards in role references
		role, err := a.Client.GetClient().RbacV1().Roles(rb.Namespace).Get(a.Context, rb.RoleRef.Name, metav1.GetOptions{})
		if err != nil {
			continue
		}

		for _, rule := range role.Rules {
			if containsWildcard(rule.Verbs) || containsWildcard(rule.Resources) {
				failures = append(failures, common.Failure{
					Text:      fmt.Sprintf("RoleBinding %s references Role %s which contains wildcard permissions - this is not recommended for security best practices", rb.Name, role.Name),
					Sensitive: []common.Sensitive{},
				})
			}
		}

		if len(failures) > 0 {
			results = append(results, common.Result{
				Kind:  "Security/RoleBinding",
				Name:  fmt.Sprintf("%s/%s", rb.Namespace, rb.Name),
				Error: failures,
			})
			AnalyzerErrorsMetric.WithLabelValues("Security/RoleBinding", rb.Name, rb.Namespace).Set(float64(len(failures)))
		}
	}

	return results, nil
}

func analyzePodSecurityContexts(a common.Analyzer) ([]common.Result, error) {
	var results []common.Result

	pods, err := a.Client.GetClient().CoreV1().Pods(a.Namespace).List(a.Context, metav1.ListOptions{
		LabelSelector: a.LabelSelector,
	})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		var failures []common.Failure

		// Check for privileged containers first (most critical)
		hasPrivilegedContainer := false
		for _, container := range pod.Spec.Containers {
			if container.SecurityContext != nil && container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
				failures = append(failures, common.Failure{
					Text:      fmt.Sprintf("Container %s in pod %s is running as privileged which poses security risks", container.Name, pod.Name),
					Sensitive: []common.Sensitive{},
				})
				hasPrivilegedContainer = true
				break
			}
		}

		// Only check for missing security context if no privileged containers found
		if !hasPrivilegedContainer && pod.Spec.SecurityContext == nil {
			failures = append(failures, common.Failure{
				Text:      fmt.Sprintf("Pod %s does not have a security context defined which may pose security risks", pod.Name),
				Sensitive: []common.Sensitive{},
			})
		}

		if len(failures) > 0 {
			results = append(results, common.Result{
				Kind:  "Security/Pod",
				Name:  fmt.Sprintf("%s/%s", pod.Namespace, pod.Name),
				Error: failures[:1],
			})
			AnalyzerErrorsMetric.WithLabelValues("Security/Pod", pod.Name, pod.Namespace).Set(1)
		}
	}

	return results, nil
}

func containsWildcard(slice []string) bool {
	for _, item := range slice {
		if item == "*" {
			return true
		}
	}
	return false
}
