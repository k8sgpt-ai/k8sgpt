package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/spf13/viper"
	"os"
)

type AWS struct {
	sess *session.Session
}

func (a *AWS) Deploy(namespace string) error {

	return nil
}

func (a *AWS) UnDeploy(namespace string) error {
	a.sess = nil
	return nil
}

func (a *AWS) AddAnalyzer(mergedMap *map[string]common.IAnalyzer) {
	// Check for AWS credentials in the environment
	// https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		panic("AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY must be set in the environment")
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            aws.Config{},
	}))
	a.sess = sess
	(*mergedMap)["EKS"] = &EKSAnalyzer{
		session: a.sess,
	}
}

func (a *AWS) GetAnalyzerName() []string {

	return []string{"EKS"}
}

func (a *AWS) GetNamespace() (string, error) {

	return "", nil
}

func (a *AWS) OwnsAnalyzer(s string) bool {
	for _, az := range a.GetAnalyzerName() {
		if s == az {
			return true
		}
	}
	return false
}

func (a *AWS) isFilterActive() bool {
	activeFilters := viper.GetStringSlice("active_filters")

	for _, filter := range a.GetAnalyzerName() {
		for _, af := range activeFilters {
			if af == filter {
				return true
			}
		}
	}

	return false
}

func (a *AWS) IsActivate() bool {
	if a.isFilterActive() {
		return true
	} else {
		return false
	}
}

func NewAWS() *AWS {
	return &AWS{}
}
