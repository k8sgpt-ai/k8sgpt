package analyzer

import (
	"fmt"
	"time"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	cron "github.com/robfig/cron/v3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CronJobAnalyzer struct{}

func (analyzer CronJobAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	var results []common.Result

	cronJobList, err := a.Client.GetClient().BatchV1().CronJobs("").List(a.Context, v1.ListOptions{})
	if err != nil {
		return results, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, cronJob := range cronJobList.Items {
		var failures []common.Failure
		if cronJob.Spec.Suspend != nil && *cronJob.Spec.Suspend {
			failures = append(failures, common.Failure{
				Text:      fmt.Sprintf("CronJob %s is suspended", cronJob.Name),
				Sensitive: []common.Sensitive{},
			})
		} else {
			// check the schedule format
			if _, err := CheckCronScheduleIsValid(cronJob.Spec.Schedule); err != nil {
				failures = append(failures, common.Failure{
					Text:      fmt.Sprintf("CronJob %s has an invalid schedule: %s", cronJob.Name, err.Error()),
					Sensitive: []common.Sensitive{},
				})
			}

			// check the starting deadline
			if cronJob.Spec.StartingDeadlineSeconds != nil {
				deadline := time.Duration(*cronJob.Spec.StartingDeadlineSeconds) * time.Second
				if deadline < 0 {

					failures = append(failures, common.Failure{
						Text:      fmt.Sprintf("CronJob %s has a negative starting deadline", cronJob.Name),
						Sensitive: []common.Sensitive{},
					})

				}
			}

		}

		if len(failures) > 0 {
			preAnalysis[cronJob.Name] = common.PreAnalysis{
				FailureDetails: failures,
			}
		}

		for key, value := range preAnalysis {
			currentAnalysis := common.Result{
				Kind:         "CronJob",
				Name:         key,
				Error:        value.FailureDetails,
				ParentObject: "",
			}
			a.Results = append(results, currentAnalysis)
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
