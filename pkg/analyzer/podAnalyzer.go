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

func AnalyzePod(ctx context.Context, client *kubernetes.Client, aiClient ai.IAI, explain bool, analysisResults *[]Analysis) error {

	// search all namespaces for pods that are not running
	list, err := client.GetClient().CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	var preAnalysis = map[string]PreAnalysis{}

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
					evt, err := FetchLatestPodEvent(ctx, client, &pod)
					if err != nil {
						continue
					}
					if evt.Reason == "FailedCreatePodSandBox" && evt.Message != "" {
						failures = append(failures, evt.Message)
					}
				}
			}
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)] = PreAnalysis{
				Pod:            pod,
				FailureDetails: failures,
			}
		}
	}

	for key, value := range preAnalysis {
		inputValue := strings.Join(value.FailureDetails, " ")
		var currentAnalysis = Analysis{
			Kind:  "Pod",
			Name:  key,
			Error: value.FailureDetails[0],
		}

		parent, _ := getParent(client, value.Pod.ObjectMeta)

		if explain {
			s := spinner.New(spinner.CharSets[35], 100*time.Millisecond) // Build our new spinner
			s.Start()

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
				currentAnalysis.Details = string(output)
				*analysisResults = append(*analysisResults, currentAnalysis)
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
			currentAnalysis.Details = response
		}
		currentAnalysis.ParentObject = parent
		*analysisResults = append(*analysisResults, currentAnalysis)
	}

	return nil
}

func getParent(client *kubernetes.Client, meta metav1.ObjectMeta) (string, bool) {
	if meta.OwnerReferences != nil {
		for _, owner := range meta.OwnerReferences {
			switch owner.Kind {
			case "ReplicaSet":
				rs, err := client.GetClient().AppsV1().ReplicaSets(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if rs.OwnerReferences != nil {
					return getParent(client, rs.ObjectMeta)
				}
				return "ReplicaSet/" + rs.Name, false

			case "Deployment":
				dep, err := client.GetClient().AppsV1().Deployments(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if dep.OwnerReferences != nil {
					return getParent(client, dep.ObjectMeta)
				}
				return "Deployment/" + dep.Name, false

			case "StatefulSet":
				sts, err := client.GetClient().AppsV1().StatefulSets(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if sts.OwnerReferences != nil {
					return getParent(client, sts.ObjectMeta)
				}
				return "StatefulSet/" + sts.Name, false

			case "DaemonSet":
				ds, err := client.GetClient().AppsV1().DaemonSets(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if ds.OwnerReferences != nil {
					return getParent(client, ds.ObjectMeta)
				}
				return "DaemonSet/" + ds.Name, false
			}
		}
	}
	return meta.Name, false
}
