/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package find

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/cloud-native-skunkworks/k8sgpt/pkg/client"
	"github.com/cloud-native-skunkworks/k8sgpt/pkg/openai"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var explain bool

// problemsCmd represents the problems command
var problemsCmd = &cobra.Command{
	Use:   "problems",
	Short: "This command will find problems within your Kubernetes cluster",
	Long: `This command will find problems within your Kubernetes cluster and
	 provide you with a list of issues that need to be resolved`,
	Run: func(cmd *cobra.Command, args []string) {

		// Initialise the openAI client
		openAIClient, err := openai.NewClient()
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		ctx := context.Background()
		// Get kubernetes client from viper
		client := viper.Get("kubernetesClient").(*client.Client)

		// search all namespaces for pods that are not running
		list, err := client.GetClient().CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		var brokenPods = map[string][]string{}

		for _, pod := range list.Items {
			// Loop through container status to check for crashes
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

				response, err := openAIClient.GetCompletion(ctx, strings.Join(value, " "))
				s.Stop()
				if err != nil {
					color.Red("Error: %v", err)
					return
				}

				color.Green(response)
			}
		}

	},
}

func init() {

	problemsCmd.Flags().BoolVarP(&explain, "explain", "e", false, "Explain the problem to me")

	FindCmd.AddCommand(problemsCmd)

}
