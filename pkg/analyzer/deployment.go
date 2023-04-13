package analyzer

import (
	"context"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
)

// DeploymentAnalyzer is an analyzer that checks for misconfigured Deployments
type DeploymentAnalyzer struct {
}

// Analyze scans all namespaces for Deployments with misconfigurations
func (d DeploymentAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	var results []common.Result
	deployments, err := a.Client.GetClient().AppsV1().Deployments("").List(context.Background(), v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, deployment := range deployments.Items {
		if *deployment.Spec.Replicas != deployment.Status.Replicas {
			failureDetails := []common.Failure{
				{
					Text:      fmt.Sprintf("Deployment %s has a mismatch between the desired and actual replicas", deployment.Name),
					Sensitive: []common.Sensitive{},
				},
			}

			result := common.Result{
				Kind:         "Deployment",
				Name:         fmt.Sprintf("%s/%s", deployment.Namespace, deployment.Name),
				Error:        failureDetails,
				ParentObject: "",
			}

			results = append(results, result)
		}
	}

	return results, nil
}
