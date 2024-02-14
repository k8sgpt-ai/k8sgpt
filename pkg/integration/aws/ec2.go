package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
)

type EC2Analyzer struct {
	session *session.Session
}

func (e *EC2Analyzer) Analyze(analysis common.Analyzer) ([]common.Result, error) {
	var common []common.Result = []common.Result{}
	// ec2session
	ec2session := ec2.New(e.session)

	// get all compute instances
	instances, err := ec2session.DescribeInstances(nil)
	if err != nil {
		return nil, err
	}

	// print each instance name
	for _, reservation := range instances.Reservations {
		for _, instance := range reservation.Instances {
			println(*instance.InstanceId)
		}
	}

	// get all instances
	return common, nil
}
