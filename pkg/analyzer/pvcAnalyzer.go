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

func AnalyzePersistentVolumeClaim(ctx context.Context, client *kubernetes.Client, aiClient ai.IAI, explain bool, analysisResults *[]Analysis) error {

	// search all namespaces for pods that are not running
	list, err := client.GetClient().CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	var preAnalysis = map[string]PreAnalysis{}

	for _, pvc := range list.Items {
		var failures []string

		// Check for empty rs
		if pvc.Status.Phase == "Pending" {

			// parse the event log and append details
			evt, err := FetchLatestPvcEvent(ctx, client, &pvc)
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
		var currentAnalysis = Analysis{
			Kind:  "PersistentVolumeClaim",
			Name:  key,
			Error: value.FailureDetails[0],
		}

		parent, _ := getParent(client, value.PersistentVolumeClaim.ObjectMeta)

		if explain {
			s := spinner.New(spinner.CharSets[35], 100*time.Millisecond) // Build our new spinner
			s.Start()

			inputValue := strings.Join(value.FailureDetails, " ")

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
				currentAnalysis.ParentObject = parent
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
