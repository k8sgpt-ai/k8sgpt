package analyzer

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceAnalyzer struct{}

func (ServiceAnalyzer) RunAnalysis(ctx context.Context, config *AnalysisConfiguration, client *kubernetes.Client, aiClient ai.IAI,
	analysisResults *[]Analysis) error {

	// search all namespaces for pods that are not running
	list, err := client.GetClient().CoreV1().Endpoints(config.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	var preAnalysis = map[string]PreAnalysis{}

	for _, ep := range list.Items {
		var failures []string

		// Check for empty service
		if len(ep.Subsets) == 0 {
			svc, err := client.GetClient().CoreV1().Services(ep.Namespace).Get(ctx, ep.Name, metav1.GetOptions{})
			if err != nil {
				color.Yellow("Service %s/%s does not exist", ep.Namespace, ep.Name)
				continue
			}

			for k, v := range svc.Spec.Selector {
				failures = append(failures, fmt.Sprintf("Service has no endpoints, expected label %s=%s", k, v))
			}
		} else {
			count := 0
			pods := []string{}

			// Check through container status to check for crashes
			for _, epSubset := range ep.Subsets {
				if len(epSubset.NotReadyAddresses) > 0 {
					for _, addresses := range epSubset.NotReadyAddresses {
						count++
						pods = append(pods, addresses.TargetRef.Kind+"/"+addresses.TargetRef.Name)
					}
					failures = append(failures, fmt.Sprintf("Service has not ready endpoints, pods: %s, expected %d", pods, count))
				}
			}
		}

		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", ep.Namespace, ep.Name)] = PreAnalysis{
				Endpoint:       ep,
				FailureDetails: failures,
			}
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = Analysis{
			Kind:  "Service",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(client, value.Endpoint.ObjectMeta)
		currentAnalysis.ParentObject = parent
		*analysisResults = append(*analysisResults, currentAnalysis)
	}
	return nil
}
