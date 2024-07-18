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
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"

	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/crd/api/policyreport/v1alpha2"
)

//	"github.com/kyverno/policy-reporter-kyverno-plugin/pkg/crd/api/policyreport/v1alpha2"

type KyvernoAnalyzer struct {
	policyReportAnalysis  bool
	clusterReportAnalysis bool
}

func (KyvernoAnalyzer) analyzePolicyReports(a common.Analyzer) ([]common.Result, error) {
	result := &v1alpha2.PolicyReportList{}
	client := a.Client.CtrlClient

	err := v1alpha2.AddToScheme(client.Scheme())
	if err != nil {
		return nil, err
	}
	if err := client.List(a.Context, result, &ctrl.ListOptions{}); err != nil {
		return nil, err
	}

	// Find criticals and get CVE
	var preAnalysis = map[string]common.PreAnalysis{}

	for _, report := range result.Items {

		// For each pod there may be multiple vulnerabilities
		var failures []common.Failure
		for _, vuln := range report.Results {
			if vuln.Result == "fail" {
				// get the vulnerability ID
				// get the vulnerability description
				failures = append(failures, common.Failure{
					Text:      fmt.Sprintf("policy failure: %s (message: %s)", vuln.Policy, vuln.Message),
					Sensitive: []common.Sensitive{},
				})
			}
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", report.Namespace,
				report.Name)] = common.PreAnalysis{
				KyvernoPolicyReport: report,
				FailureDetails:      failures,
			}
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  "PolicyReport",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.KyvernoPolicyReport.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil

}

func (t KyvernoAnalyzer) analyzeClusterPolicyReports(a common.Analyzer) ([]common.Result, error) {
	result := &v1alpha2.ClusterPolicyReportList{}
	client := a.Client.CtrlClient

	err := v1alpha2.AddToScheme(client.Scheme())
	if err != nil {
		return nil, err
	}
	if err := client.List(a.Context, result, &ctrl.ListOptions{}); err != nil {
		return nil, err
	}

	// Find criticals and get CVE
	var preAnalysis = map[string]common.PreAnalysis{}

	for _, report := range result.Items {

		// For each pod there may be multiple vulnerabilities
		var failures []common.Failure
		for _, vuln := range report.Results {
			if vuln.Severity == "CRITICAL" {
				// get the vulnerability ID
				// get the vulnerability description
				failures = append(failures, common.Failure{
					Text:      fmt.Sprintf("critical Vulnerability found ID: %s (learn more at: %s)", vuln.ID, vuln.Source),
					Sensitive: []common.Sensitive{},
				})
			}
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", report.Namespace,
				report.Name)] = common.PreAnalysis{
				KyvernoClusterPolicyReport: report,
				FailureDetails:             failures,
			}
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  "ClusterPolicyReport",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.KyvernoClusterPolicyReport.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}

func (t KyvernoAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	if t.policyReportAnalysis {
		common := make([]common.Result, 0)
		vresult, err := t.analyzePolicyReports(a)
		if err != nil {
			return nil, err
		}
		common = append(common, vresult...)
		return common, nil
	}
	if t.clusterReportAnalysis {
		common := make([]common.Result, 0)
		cresult, err := t.analyzeClusterPolicyReports(a)
		if err != nil {
			return nil, err
		}
		common = append(common, cresult...)
		return common, nil
	}
	return make([]common.Result, 0), nil
}
