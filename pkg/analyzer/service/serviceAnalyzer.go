package service

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceAnalyzer struct{}

func (ServiceAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	// search all namespaces for pods that are not running
	list, err := a.Client.GetClient().CoreV1().Endpoints(a.Namespace).List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, ep := range list.Items {
		var failures []string

		// Check for empty service
		if len(ep.Subsets) == 0 {
			svc, err := a.Client.GetClient().CoreV1().Services(ep.Namespace).Get(a.Context, ep.Name, metav1.GetOptions{})
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
			preAnalysis[fmt.Sprintf("%s/%s", ep.Namespace, ep.Name)] = common.PreAnalysis{
				Endpoint:       ep,
				FailureDetails: failures,
			}
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  "Service",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.Endpoint.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}
	return a.Results, nil
}
