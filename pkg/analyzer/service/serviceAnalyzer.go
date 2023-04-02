package service

import (
	"fmt"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/common"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceAnalyzer struct {
	common.Analyzer
}

func (a *ServiceAnalyzer) Analyze() error {
	// search all namespaces for pods that are not running
	list, err := a.Client.GetClient().CoreV1().Endpoints(a.Namespace).List(a.Context, metav1.ListOptions{})
	if err != nil {
		return err
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
		a.Result = append(a.Result, currentAnalysis)
	}
	return nil
}

func (a *ServiceAnalyzer) GetResult() []common.Result {
	return a.Result
}
