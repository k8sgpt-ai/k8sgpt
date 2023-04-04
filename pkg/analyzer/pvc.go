package analyzer

import (
	"fmt"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PvcAnalyzer struct{}

func (PvcAnalyzer) Analyze(a Analyzer) ([]Result, error) {

	// search all namespaces for pods that are not running
	list, err := a.Client.GetClient().CoreV1().PersistentVolumeClaims(a.Namespace).List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]PreAnalysis{}

	for _, pvc := range list.Items {
		var failures []string

		// Check for empty rs
		if pvc.Status.Phase == "Pending" {

			// parse the event log and append details
			evt, err := FetchLatestEvent(a.Context, a.Client, pvc.Namespace, pvc.Name)
			if err != nil || evt == nil {
				continue
			}
			if evt.Reason == "ProvisioningFailed" && evt.Message != "" {
				failures = append(failures, evt.Message)
			}
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", pvc.Namespace, pvc.Name)] = PreAnalysis{
				PersistentVolumeClaim: pvc,
				FailureDetails:        failures,
			}
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = Result{
			Kind:  "PersistentVolumeClaim",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.PersistentVolumeClaim.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}
