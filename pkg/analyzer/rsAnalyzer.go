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

func AnalyzeReplicaSet(ctx context.Context, client *kubernetes.Client, aiClient ai.IAI, explain bool, analysisResults *[]Analysis) error {

	// search all namespaces for pods that are not running
	list, err := client.GetClient().AppsV1().ReplicaSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	var brokenRS = map[string][]string{}

	for _, rs := range list.Items {

		// Check for empty rs
		if rs.Status.Replicas == 0 {

			// Check through container status to check for crashes
			for _, rsStatus := range rs.Status.Conditions {
				if rsStatus.Type == "ReplicaFailure" && rsStatus.Reason == "FailedCreate" {
					brokenRS[fmt.Sprintf("%s/%s", rs.Namespace, rs.Name)] = []string{rsStatus.Message}
				}
			}
		}
	}

	for key, value := range brokenRS {
		var currentAnalysis = Analysis{
			Kind:  "ReplicaSet",
			Name:  key,
			Error: value[0],
		}

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
		*analysisResults = append(*analysisResults, currentAnalysis)
	}

	return nil
}
