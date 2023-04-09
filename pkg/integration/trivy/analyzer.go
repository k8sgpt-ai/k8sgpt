package trivy

import (
	"github.com/aquasecurity/trivy-operator/pkg/apis/aquasecurity/v1alpha1"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
)

type TrivyAnalyzer struct {
}

func (TrivyAnalyzer) Analyze(a analyzer.Analyzer) ([]analyzer.Result, error) {

	// Get all trivy VulnerabilityReports
	result := &v1alpha1.VulnerabilityReportList{}

	err := a.Client.GetRestClient().Get().Namespace(a.Namespace).Resource("vulnerabilityreports").Do(a.Context).Into(result)
	if err != nil {
		return nil, err
	}

	// WIP

	return nil, nil
}
