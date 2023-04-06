package analyzer

import (
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StatefulSetAnalyzer struct{}

func (StatefulSetAnalyzer) Analyze(a Analyzer) ([]Result, error) {
	list, err := a.Client.GetClient().AppsV1().StatefulSets(a.Namespace).List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var preAnalysis = map[string]PreAnalysis{}

	for _, sts := range list.Items {
		var failures []string

		// get serviceName
		serviceName := sts.Spec.ServiceName
		_, err := a.Client.GetClient().CoreV1().Services(sts.Namespace).Get(a.Context, serviceName, metav1.GetOptions{})
		if err != nil {
			failures = append(failures, fmt.Sprintf("StatefulSet uses the service %s/%s which does not exist.", sts.Namespace, serviceName))
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", sts.Namespace, sts.Name)] = PreAnalysis{
				StatefulSet:    sts,
				FailureDetails: failures,
			}
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = Result{
			Kind:  "StatefulSet",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.StatefulSet.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}
