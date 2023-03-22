package analyzer

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/spf13/viper"
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
				// This represents a container that is still being created or blocked due to conditions such as OOMKilled
				if containerStatus.State.Waiting.Reason == "ContainerCreating" && pod.Status.Phase == "Pending" {

					// parse the event log and append details
					evt, err := FetchLatestPodEvent(ctx, client, &pod)
					if err != nil {
						continue
					}
					if evt.Reason == "FailedCreatePodSandBox" {
						failureDetails = append(failureDetails, evt.Message)
						brokenPods[fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)] = failureDetails
					}
				}

			}
		}

	}

	count := 0
	for key, value := range brokenPods {
		fmt.Printf("%s: %s: %s\n", color.CyanString("%d", count), color.YellowString(key), color.RedString(value[0]))
		count++
		if explain {
			s := spinner.New(spinner.CharSets[35], 100*time.Millisecond) // Build our new spinner
			s.Start()

			inputValue := strings.Join(value, " ")

			// Check for cached data
			sEnc := base64.StdEncoding.EncodeToString([]byte(inputValue))
			// find in viper cache
			if viper.IsSet(sEnc) {
				s.Stop()
				// retrieve data from cache
				response := viper.GetString(sEnc)
				if response == "" {
					color.Red("error retrieving cached data")
					continue
				}
				output, err := base64.StdEncoding.DecodeString(response)
				if err != nil {
					color.Red("error decoding cached data: %v", err)
					continue
				}

				color.Green(string(output))
				continue
			}

			response, err := aiClient.GetCompletion(ctx, inputValue)
			s.Stop()
			if err != nil {
				color.Red("error getting completion: %v", err)
				continue
			}

			if !viper.IsSet(sEnc) {
				viper.Set(sEnc, base64.StdEncoding.EncodeToString([]byte(response)))
				if err := viper.WriteConfig(); err != nil {
					return err
				}
			}

			color.Green(response)
		}
	}

	return nil
}
