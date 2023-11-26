/*
Copyright 2023 The K8sGPT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package analyzer

import (
	"fmt"
	"github.com/spf13/viper"
	"time"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	cron "github.com/robfig/cron/v3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type CronJobAnalyzer struct{}

func (analyzer CronJobAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "CronJob"
	apiDoc := kubernetes.K8sApiReference{
		Kind: kind,
		ApiVersion: schema.GroupVersion{
			Group:   "batch",
			Version: "v1",
		},
		OpenapiSchema: a.OpenapiSchema,
	}

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	// get analyzer flags
	cronjobMaxTime := viper.GetString("cronjobMaxTime")

	cronJobList, err := a.Client.GetClient().BatchV1().CronJobs(a.Namespace).List(a.Context, v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, cronJob := range cronJobList.Items {
		var failures []common.Failure
		if cronJob.Spec.Suspend != nil && *cronJob.Spec.Suspend {
			doc := apiDoc.GetApiDocV2("spec.suspend")

			failures = append(failures, common.Failure{
				Text:          fmt.Sprintf("CronJob %s is suspended", cronJob.Name),
				KubernetesDoc: doc,
				Sensitive: []common.Sensitive{
					{
						Unmasked: cronJob.Namespace,
						Masked:   util.MaskString(cronJob.Namespace),
					},
					{
						Unmasked: cronJob.Name,
						Masked:   util.MaskString(cronJob.Name),
					},
				},
			})
		} else {
			// check the schedule format
			if _, err := CheckCronScheduleIsValid(cronJob.Spec.Schedule); err != nil {
				doc := apiDoc.GetApiDocV2("spec.schedule")

				failures = append(failures, common.Failure{
					Text:          fmt.Sprintf("CronJob %s has an invalid schedule: %s", cronJob.Name, err.Error()),
					KubernetesDoc: doc,
					Sensitive: []common.Sensitive{
						{
							Unmasked: cronJob.Namespace,
							Masked:   util.MaskString(cronJob.Namespace),
						},
						{
							Unmasked: cronJob.Name,
							Masked:   util.MaskString(cronJob.Name),
						},
					},
				})
			}

			// check the starting deadline
			if cronJob.Spec.StartingDeadlineSeconds != nil {
				deadline := time.Duration(*cronJob.Spec.StartingDeadlineSeconds) * time.Second
				if deadline < 0 {
					doc := apiDoc.GetApiDocV2("spec.startingDeadlineSeconds")

					failures = append(failures, common.Failure{
						Text:          fmt.Sprintf("CronJob %s has a negative starting deadline", cronJob.Name),
						KubernetesDoc: doc,
						Sensitive: []common.Sensitive{
							{
								Unmasked: cronJob.Namespace,
								Masked:   util.MaskString(cronJob.Namespace),
							},
							{
								Unmasked: cronJob.Name,
								Masked:   util.MaskString(cronJob.Name),
							},
						},
					})

				}
			}

			// check cronjob running time is exceed cronjobMaxTime
			if cronjobMaxTime != "" && cronJob.Status.Active != nil && len(cronJob.Status.Active) > 0 {
				// parse cronjobMaxTime
				maxTimeDuration, err := time.ParseDuration(cronjobMaxTime)
				if err != nil {
					return nil, err
				}
				jobName := cronJob.Status.Active[0].Name

				// get running pod of cronjob
				pods, err := a.Client.GetClient().CoreV1().Pods(a.Namespace).List(a.Context, v1.ListOptions{
					LabelSelector: fmt.Sprintf("job-name=%s", jobName),
				})
				if err != nil {
					return nil, err
				}
				if len(pods.Items) > 0 {
					pod := pods.Items[0]
					if pod.Status.StartTime != nil {
						runningTime := time.Since(pod.Status.StartTime.Time)
						if runningTime > maxTimeDuration {
							doc := apiDoc.GetApiDocV2("spec.status")

							failures = append(failures, common.Failure{
								Text:          fmt.Sprintf("CronJob %s is running for more than max time threshold: %s", cronJob.Name, cronjobMaxTime),
								KubernetesDoc: doc,
								Sensitive: []common.Sensitive{
									{
										Unmasked: cronJob.Namespace,
										Masked:   util.MaskString(cronJob.Namespace),
									},
									{
										Unmasked: cronJob.Name,
										Masked:   util.MaskString(cronJob.Name),
									},
								},
							})
						}
					}
				}
			}
		}

		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", cronJob.Namespace, cronJob.Name)] = common.PreAnalysis{
				FailureDetails: failures,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, cronJob.Name, cronJob.Namespace).Set(float64(len(failures)))

		}

		for key, value := range preAnalysis {
			currentAnalysis := common.Result{
				Kind:  kind,
				Name:  key,
				Error: value.FailureDetails,
			}
			a.Results = append(a.Results, currentAnalysis)
		}
	}

	return a.Results, nil
}

// Check CRON schedule format
func CheckCronScheduleIsValid(schedule string) (bool, error) {
	_, err := cron.ParseStandard(schedule)
	if err != nil {
		return false, err
	}

	return true, nil
}
