package analyzer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RunAnalysis(ctx context.Context, client *kubernetes.Client, aiClient *ai.Client, explain bool) error {

	// search all namespaces for pods that are not running
	list, err := client.GetClient().CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	var brokenPods = map[string][]string{}

	for _, pod := range list.Items {

		// Check for pending pods
		if pod.Status.Phase == "Pending" {

			// Check through container status to check for crashes
			for _, containerStatus := range pod.Status.Conditions {
				if containerStatus.Type == "PodScheduled" && containerStatus.Reason == "Unschedulable" {
					brokenPods[fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)] = []string{containerStatus.Message}
				}
			}
		}

		// Check through container status to check for crashes
		var failureDetails = []string{}
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil {
				if containerStatus.State.Waiting.Reason == "CrashLoopBackOff" || containerStatus.State.Waiting.Reason == "ImagePullBackOff" {

					failureDetails = append(failureDetails, containerStatus.State.Waiting.Message)
					brokenPods[fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)] = failureDetails
				}

			}
		}

	}
	for key, value := range brokenPods {
		fmt.Printf("%s: %s\n", color.YellowString(key), color.RedString(value[0]))

		if explain {
			s := spinner.New(spinner.CharSets[35], 100*time.Millisecond) // Build our new spinner
			s.Start()

			response, err := aiClient.GetCompletion(ctx, strings.Join(value, " "))
			s.Stop()
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", color.GreenString(response))
		}
	}

	return nil
}
