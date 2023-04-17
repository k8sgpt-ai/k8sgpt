package analyzer

import (
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HpaAnalyzer struct{}

func (HpaAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "HorizontalPodAutoscaler"

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	list, err := a.Client.GetClient().AutoscalingV1().HorizontalPodAutoscalers(a.Namespace).List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, hpa := range list.Items {
		var failures []common.Failure

		// check ScaleTargetRef exist
		scaleTargetRef := hpa.Spec.ScaleTargetRef
		scaleTargetRefNotFound := false

		switch scaleTargetRef.Kind {
		case "Deployment":
            deployment, err := a.Client.GetClient().AppsV1().Deployments(hpa.Namespace).Get(a.Context, scaleTargetRef.Name, metav1.GetOptions{})
            if err != nil {
                scaleTargetRefNotFound = true
            } else {
                // check if the deployment has resource configured
                if (deployment.Spec.Template.Spec.Containers[0].Resources.Requests == nil) || (deployment.Spec.Template.Spec.Containers[0].Resources.Limits == nil) {
                    failures = append(failures, common.Failure{
                        Text: fmt.Sprintf("Deployment %s/%s does not have resource configured.", deployment.Namespace, deployment.Name),
                        Sensitive: []common.Sensitive{
                            {
                                Unmasked: deployment.Name,
                                Masked: util.MaskString(deployment.Name),
                            },
                        },
                    })
                }
            }
		case "ReplicationController":
			_, err := a.Client.GetClient().CoreV1().ReplicationControllers(hpa.Namespace).Get(a.Context, scaleTargetRef.Name, metav1.GetOptions{})
			if err != nil {
				scaleTargetRefNotFound = true
			}
		case "ReplicaSet":
			_, err := a.Client.GetClient().AppsV1().ReplicaSets(hpa.Namespace).Get(a.Context, scaleTargetRef.Name, metav1.GetOptions{})
			if err != nil {
				scaleTargetRefNotFound = true
			}
		case "StatefulSet":
			_, err := a.Client.GetClient().AppsV1().StatefulSets(hpa.Namespace).Get(a.Context, scaleTargetRef.Name, metav1.GetOptions{})
			if err != nil {
				scaleTargetRefNotFound = true
			}
		default:
			failures = append(failures, common.Failure{
				Text:      fmt.Sprintf("HorizontalPodAutoscaler uses %s as ScaleTargetRef which is not an option.", scaleTargetRef.Kind),
				Sensitive: []common.Sensitive{},
			})
		}

		if scaleTargetRefNotFound {
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf("HorizontalPodAutoscaler uses %s/%s as ScaleTargetRef which does not exist.", scaleTargetRef.Kind, scaleTargetRef.Name),
				Sensitive: []common.Sensitive{
					{
						Unmasked: scaleTargetRef.Name,
						Masked:   util.MaskString(scaleTargetRef.Name),
					},
				},
			})
		}

		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", hpa.Namespace, hpa.Name)] = common.PreAnalysis{
				HorizontalPodAutoscalers: hpa,
				FailureDetails:           failures,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, hpa.Name, hpa.Namespace).Set(float64(len(failures)))
		}

	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.HorizontalPodAutoscalers.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}
