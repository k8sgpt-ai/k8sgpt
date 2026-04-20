package analyzer

import (
	"context"
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type DaemonSetAnalyzer struct {
}

func (DaemonSetAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	kind := "DaemonSet"
	apiDoc := kubernetes.K8sApiReference{
		Kind: kind,
		ApiVersion: schema.GroupVersion{
			Group:   "apps",
			Version: "v1",
		},
		OpenapiSchema: a.OpenapiSchema,
	}

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	daemonsets, err := a.Client.GetClient().AppsV1().DaemonSets(a.Namespace).List(a.Context, metav1.ListOptions{LabelSelector: a.LabelSelector})
	if err != nil {
		return nil, err
	}
	var preAnalysis = map[string]common.PreAnalysis{}

	for _, ds := range daemonsets.Items {
		var failures []common.Failure

		if ds.Status.NumberReady < ds.Status.DesiredNumberScheduled {
			doc := apiDoc.GetApiDocV2("status.numberReady")
			failures = append(failures, common.Failure{
				Text:          fmt.Sprintf("DaemonSet %s/%s has %d ready replicas out of %d scheduled", ds.Namespace, ds.Name, ds.Status.NumberReady, ds.Status.DesiredNumberScheduled),
				KubernetesDoc: doc,
				Sensitive: []common.Sensitive{
					{
						Unmasked: ds.Namespace,
						Masked:   util.MaskString(ds.Namespace),
					},
					{
						Unmasked: ds.Name,
						Masked:   util.MaskString(ds.Name),
					},
				},
			})
		}
		for _, imagePullSecret := range ds.Spec.Template.Spec.ImagePullSecrets {
			_, err := a.Client.GetClient().CoreV1().Secrets(ds.Namespace).Get(context.Background(), imagePullSecret.Name, metav1.GetOptions{})
			if err != nil {
				failures = append(failures, common.Failure{
					Text: fmt.Sprintf("DaemonSet %s/%s references non-existent ImagePullSecret %s", ds.Namespace, ds.Name, imagePullSecret.Name),
					Sensitive: []common.Sensitive{
						{
							Unmasked: ds.Namespace,
							Masked:   util.MaskString(ds.Namespace),
						},
						{
							Unmasked: ds.Name,
							Masked:   util.MaskString(ds.Name),
						},
						{
							Unmasked: imagePullSecret.Name,
							Masked:   util.MaskString(imagePullSecret.Name),
						},
					},
				})
			}
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", ds.Namespace, ds.Name)] = common.PreAnalysis{
				FailureDetails: failures,
				DaemonSet:      ds,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, ds.Name, ds.Namespace).Set(float64(len(failures)))
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, found := util.GetParent(a.Client, value.DaemonSet.ObjectMeta)
		if found {
			currentAnalysis.ParentObject = parent
		}
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}
