package trivy

import (
	"fmt"

	"github.com/aquasecurity/trivy-operator/pkg/apis/aquasecurity/v1alpha1"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
)

type TrivyAnalyzer struct {
}

func (TrivyAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	// Get all trivy VulnerabilityReports
	result := &v1alpha1.VulnerabilityReportList{}

	err := a.Client.GetRestClient().Get().Namespace(a.Namespace).Resource("vulnerabilityreports").Do(a.Context).Into(result)
	if err != nil {
		return nil, err
	}

	// Find criticals and get CVE
	var preAnalysis = map[string]common.PreAnalysis{}

	for _, report := range result.Items {

		// For each pod there may be multiple vulnerabilities
		var failures []string
		for _, vuln := range report.Report.Vulnerabilities {
			if vuln.Severity == "CRITICAL" {
				// get the vulnerability ID
				// get the vulnerability description
				failures = append(failures, fmt.Sprintf("Critical Vulnerability found ID: %s, Description: %s", vuln.VulnerabilityID, vuln.Description))
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
