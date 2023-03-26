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

func AnalyzeEndpoints(ctx context.Context, client *kubernetes.Client, aiClient ai.IAI, explain bool, analysisResults *[]Analysis) error {

	// search all namespaces for pods that are not running
	list, err := client.GetClient().CoreV1().Endpoints("").List(ctx, metav1.ListOptions{})
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
				return err
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
			Error: value.FailureDetails[0],
		}

		parent, _ := getParent(client, value.Endpoint.ObjectMeta)

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
