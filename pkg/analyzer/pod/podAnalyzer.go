package pod

import (
	"fmt"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodAnalyzer struct {
}

func (PodAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	// search all namespaces for pods that are not running
	list, err := a.Client.GetClient().CoreV1().Pods(a.Namespace).List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var preAnalysis = map[string]common.PreAnalysis{}

	for _, pod := range list.Items {
		var failures []string
		// Check for pending pods
		if pod.Status.Phase == "Pending" {

			// Check through container status to check for crashes
			for _, containerStatus := range pod.Status.Conditions {
				if containerStatus.Type == "PodScheduled" && containerStatus.Reason == "Unschedulable" {
					if containerStatus.Message != "" {
						failures = []string{containerStatus.Message}
					}
				}
			}
		}

		// Check through container status to check for crashes
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil {
				if containerStatus.State.Waiting.Reason == "CrashLoopBackOff" || containerStatus.State.Waiting.Reason == "ImagePullBackOff" {
					if containerStatus.State.Waiting.Message != "" {
						failures = append(failures, containerStatus.State.Waiting.Message)
					}
				}
				// This represents a container that is still being created or blocked due to conditions such as OOMKilled
				if containerStatus.State.Waiting.Reason == "ContainerCreating" && pod.Status.Phase == "Pending" {

					// parse the event log and append details
					evt, err := common.FetchLatestEvent(a.Context, a.Client, pod.Namespace, pod.Name)
					if err != nil || evt == nil {
						continue
					}
					if evt.Reason == "FailedCreatePodSandBox" && evt.Message != "" {
						failures = append(failures, evt.Message)
					}
				}
			}
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)] = common.PreAnalysis{
				Pod:            pod,
				FailureDetails: failures,
			}
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  "Pod",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.Pod.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}
