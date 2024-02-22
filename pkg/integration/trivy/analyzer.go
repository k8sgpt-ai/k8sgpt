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

package trivy

import (
	"fmt"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aquasecurity/trivy-operator/pkg/apis/aquasecurity/v1alpha1"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
)

type TrivyAnalyzer struct {
	vulernabilityReportAnalysis bool
	configAuditReportAnalysis   bool
}

func (TrivyAnalyzer) analyzeVulnerabilityReports(a common.Analyzer) ([]common.Result, error) {
	// Get all trivy VulnerabilityReports
	result := &v1alpha1.VulnerabilityReportList{}

	client := a.Client.CtrlClient
	err := v1alpha1.AddToScheme(client.Scheme())
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
		for _, vuln := range report.Report.Vulnerabilities {
			if vuln.Severity == "CRITICAL" {
				// get the vulnerability ID
				// get the vulnerability description
				failures = append(failures, common.Failure{
					Text:      fmt.Sprintf("critical Vulnerability found ID: %s (learn more at: %s)", vuln.VulnerabilityID, vuln.PrimaryLink),
					Sensitive: []common.Sensitive{},
				})
			}
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", report.Namespace,
				report.Name)] = common.PreAnalysis{
				TrivyVulnerabilityReport: report,
				FailureDetails:           failures,
			}
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  "VulnerabilityReport",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.TrivyVulnerabilityReport.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil

}

func (t TrivyAnalyzer) analyzeConfigAuditReports(a common.Analyzer) ([]common.Result, error) {
	// Get all trivy ConfigAuditReports
	result := &v1alpha1.ConfigAuditReportList{}

	client := a.Client.CtrlClient
	err := v1alpha1.AddToScheme(client.Scheme())
	if err != nil {
		return nil, err
	}
	if err := client.List(a.Context, result, &ctrl.ListOptions{}); err != nil {
		return nil, err
	}

	// Find criticals and get CVE
	var preAnalysis = map[string]common.PreAnalysis{}

	for _, report := range result.Items {

		// For each k8s resources there may be multiple checks
		var failures []common.Failure
		for _, check := range report.Report.Checks {
			if check.Severity == "MEDIUM" || check.Severity == "HIGH" || check.Severity == "CRITICAL" {
				failures = append(failures, common.Failure{
					Text: fmt.Sprintf("Config issue with severity \"%s\" found: %s", check.Severity, strings.Join(check.Messages, "")),
					Sensitive: []common.Sensitive{
						{
							Unmasked: report.Labels["trivy-operator.resource.name"],
							Masked:   util.MaskString(report.Labels["trivy-operator.resource.name"]),
						},
						{
							Unmasked: report.Labels["trivy-operator.resource.namespace"],
							Masked:   util.MaskString(report.Labels["trivy-operator.resource.namespace"]),
						},
					},
				})
			}
		}

		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", report.Namespace,
				report.Name)] = common.PreAnalysis{
				TrivyConfigAuditReport: report,
				FailureDetails:         failures,
			}
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  "ConfigAuditReport",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.TrivyConfigAuditReport.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}

func (t TrivyAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	if t.vulernabilityReportAnalysis {
		common := make([]common.Result, 0)
		vresult, err := t.analyzeVulnerabilityReports(a)
		if err != nil {
			return nil, err
		}
		common = append(common, vresult...)
		return common, nil
	}
	if t.configAuditReportAnalysis {
		common := make([]common.Result, 0)
		cresult, err := t.analyzeConfigAuditReports(a)
		if err != nil {
			return nil, err
		}
		common = append(common, cresult...)
		return common, nil
	}
	return make([]common.Result, 0), nil
}
