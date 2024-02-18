package aws

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"k8s.io/client-go/tools/clientcmd"
)

type EKSAnalyzer struct {
	session *session.Session
}

func (e *EKSAnalyzer) Analyze(analysis common.Analyzer) ([]common.Result, error) {
	var cr []common.Result = []common.Result{}
	_ = map[string]common.PreAnalysis{}
	svc := eks.New(e.session)
	// Get the name of the current cluster
	kubecConfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubecConfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: "",
		}).RawConfig()
	currentConfig := config.CurrentContext

	if !strings.Contains(currentConfig, "eks") {
		return cr, errors.New("EKS cluster was not detected")
	}

	input := &eks.ListClustersInput{}
	result, err := svc.ListClusters(input)
	if err != nil {
		return cr, err
	}
	for _, cluster := range result.Clusters {
		// describe the cluster
		if !strings.Contains(currentConfig, *cluster) {
			continue
		}
		input := &eks.DescribeClusterInput{
			Name: cluster,
		}
		result, err := svc.DescribeCluster(input)
		if err != nil {
			return cr, err
		}
		if len(result.Cluster.Health.Issues) > 0 {
			for _, issue := range result.Cluster.Health.Issues {
				var err = make([]common.Failure, 0)
				err = append(err, common.Failure{
					Text:          issue.String(),
					KubernetesDoc: "",
					Sensitive:     nil,
				})
				cr = append(cr, common.Result{
					Kind:  "EKS",
					Name:  "AWS/EKS",
					Error: err,
				})
			}
		}
	}
	return cr, nil
}
