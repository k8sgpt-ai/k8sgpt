/*
Copyright 2025 The K8sGPT Authors.
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

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type JobAnalyzer struct{}

func (analyzer JobAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "Job"
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

	JobList, err := a.Client.GetClient().BatchV1().Jobs(a.Namespace).List(a.Context, v1.ListOptions{LabelSelector: a.LabelSelector})
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, Job := range JobList.Items {
		var failures []common.Failure
		if Job.Spec.Suspend != nil && *Job.Spec.Suspend {
			doc := apiDoc.GetApiDocV2("spec.suspend")

			failures = append(failures, common.Failure{
				Text:          fmt.Sprintf("Job %s is suspended", Job.Name),
				KubernetesDoc: doc,
				Sensitive: []common.Sensitive{
					{
						Unmasked: Job.Namespace,
						Masked:   util.MaskString(Job.Namespace),
					},
					{
						Unmasked: Job.Name,
						Masked:   util.MaskString(Job.Name),
					},
				},
			})
		}
		if Job.Status.Failed > 0 {
			doc := apiDoc.GetApiDocV2("status.failed")
			failures = append(failures, common.Failure{
				Text:          fmt.Sprintf("Job %s has failed", Job.Name),
				KubernetesDoc: doc,
				Sensitive: []common.Sensitive{
					{
						Unmasked: Job.Namespace,
						Masked:   util.MaskString(Job.Namespace),
					},
					{
						Unmasked: Job.Name,
						Masked:   util.MaskString(Job.Name),
					},
				},
			})
		}

		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", Job.Namespace, Job.Name)] = common.PreAnalysis{
				FailureDetails: failures,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, Job.Name, Job.Namespace).Set(float64(len(failures)))
		}
	}

	for key, value := range preAnalysis {
		currentAnalysis := common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}
