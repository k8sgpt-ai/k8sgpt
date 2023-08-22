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

	"github.com/aquasecurity/trivy-operator/pkg/apis/aquasecurity/v1alpha1"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	"k8s.io/client-go/rest"
)

type TrivyAnalyzer struct {
	vulernabilityReportAnalysis bool
	configAuditReportAnalysis   bool
}

func (TrivyAnalyzer) analyzeVulnerabilityReports(a common.Analyzer) ([]common.Result, error) {
	// Get all trivy VulnerabilityReports
	result := &v1alpha1.VulnerabilityReportList{}

	config := a.Client.GetConfig()
	// Add group version to sceheme
	config.ContentConfig.GroupVersion = &v1alpha1.SchemeGroupVersion
	config.UserAgent = rest.DefaultKubernetesUserAgent()
	config.APIPath = "/apis"

	restClient, err := rest.UnversionedRESTClientFor(config)
	if err != nil {
		return nil, err
	}
	err = restClient.Get().Resource("vulnerabilityreports").Do(a.Context).Into(result)
	if err != nil {
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
			preAnalysis[fmt.Sprintf("%s/%s", report.Labels["trivy-operator.resource.namespace"],
				report.Labels["trivy-operator.resource.name"])] = common.PreAnalysis{
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
	// Get all trivy VulnerabilityReports
	result := &v1alpha1.ConfigAuditReportList{}

	config := a.Client.GetConfig()
	// Add group version to sceheme
	config.ContentConfig.GroupVersion = &v1alpha1.SchemeGroupVersion
	config.UserAgent = rest.DefaultKubernetesUserAgent()
	config.APIPath = "/apis"

	restClient, err := rest.UnversionedRESTClientFor(config)
	if err != nil {
		return nil, err
	}
	err = restClient.Get().Resource("configauditreports").Do(a.Context).Into(result)
	if err != nil {
		return nil, err
	}

	// Find criticals and get CVE
	var preAnalysis = map[string]common.PreAnalysis{}

	for _, report := range result.Items {

		var failures []common.Failure
		if report.Report.Summary.HighCount > 0 {

			failures = append(failures, common.Failure{
				Text:      fmt.Sprintf("Config audit report %s detected at least one high issue", report.Name),
				Sensitive: []common.Sensitive{},
			})

		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", report.Labels["trivy-operator.resource.namespace"],
				report.Labels["trivy-operator.resource.name"])] = common.PreAnalysis{
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
