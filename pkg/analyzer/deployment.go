package analyzer

import (
	"context"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
)

// DeploymentAnalyzer is an analyzer that checks for misconfigured Deployments
type DeploymentAnalyzer struct {
}

// Analyze scans all namespaces for Deployments with misconfigurations
func (d DeploymentAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "Deployment"

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	deployments, err := a.Client.GetClient().AppsV1().Deployments(a.Namespace).List(context.Background(), v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var preAnalysis = map[string]common.PreAnalysis{}

	for _, deployment := range deployments.Items {
		var failures []common.Failure
		if *deployment.Spec.Replicas != deployment.Status.Replicas {
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf("Deployment %s/%s has %d replicas but %d are available", deployment.Namespace, deployment.Name, *deployment.Spec.Replicas, deployment.Status.Replicas),
				Sensitive: []common.Sensitive{
					{
						Unmasked: deployment.Namespace,
						Masked:   util.MaskString(deployment.Namespace),
					},
					{
						Unmasked: deployment.Name,
						Masked:   util.MaskString(deployment.Name),
					},
				}})
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", deployment.Namespace, deployment.Name)] = common.PreAnalysis{
				FailureDetails: failures,
				Deployment:     deployment,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, deployment.Name, deployment.Namespace).Set(float64(len(failures)))
		}

	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}

		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}
