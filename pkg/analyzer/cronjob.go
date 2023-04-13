package analyzer

import (
	"fmt"
	"time"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	cron "github.com/robfig/cron/v3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CronJobAnalyzer struct{}

func (analyzer CronJobAnalyzer) Analyze(config common.Analyzer) ([]common.Result, error) {
	var results []common.Result

	cronJobList, err := config.Client.GetClient().BatchV1().CronJobs("").List(config.Context, v1.ListOptions{})
	if err != nil {
		return results, err
	}

	for _, cronJob := range cronJobList.Items {
		result := common.Result{
			Kind: "CronJob",
			Name: cronJob.Name,
		}

		if cronJob.Spec.Suspend != nil && *cronJob.Spec.Suspend {
			result.Error = append(result.Error, common.Failure{
				Text:      fmt.Sprintf("CronJob %s is suspended", cronJob.Name),
				Sensitive: []common.Sensitive{},
			})
		} else {
			// check the schedule format
			if _, err := CheckCronScheduleIsValid(cronJob.Spec.Schedule); err != nil {
				result.Error = append(result.Error, common.Failure{
					Text:      fmt.Sprintf("CronJob %s has an invalid schedule: %s", cronJob.Name, cronJob.Spec.Schedule),
					Sensitive: []common.Sensitive{},
				})
			}

			// check the starting deadline
			if cronJob.Spec.StartingDeadlineSeconds != nil {
				deadline := time.Duration(*cronJob.Spec.StartingDeadlineSeconds) * time.Second
				if deadline < 0 {

					result = common.Result{
						Kind: "CronJob",
						Name: cronJob.Name,
						Error: []common.Failure{
							{
								Text:      fmt.Sprintf("CronJob %s has a negative starting deadline: %d seconds", cronJob.Name, *cronJob.Spec.StartingDeadlineSeconds),
								Sensitive: []common.Sensitive{},
							},
						},
					}

				}
			}

		}
		results = append(results, result)
	}

	return results, nil
}

// Check CRON schedule format
func CheckCronScheduleIsValid(schedule string) (bool, error) {
	_, err := cron.ParseStandard(schedule)
	if err != nil {
		return false, err
	}

	return true, nil
}
