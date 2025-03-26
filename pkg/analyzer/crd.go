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

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	kind                  = "CustomResourceDefinition"
	webhook               = "Webhook"
	serviceNotFound       = "Custom Resource Definition Conversion Webhook Service %s not found"
	apiSpecWebhookService = "spec.conversion.webhook.clientConfig.service"
)

type CrdAnalyzer struct {
}

func (CrdAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

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

	var preAnalysis = map[string]common.PreAnalysis{}

	// Fetch all CRD's
	client := a.Client.CtrlClient
	crdList := &apiextensionsv1.CustomResourceDefinitionList{}
	client.List(a.Context, &apiextensionsv1.CustomResourceDefinitionList{})
	if err := client.List(a.Context, crdList, &ctrl.ListOptions{}); err != nil {
		return nil, err
	}

	var failures []common.Failure

	for _, crd := range crdList.Items {

		// Check crd conversion webhook service
		conversion := crd.Spec.Conversion
		if conversion.Strategy == webhook && conversion.Webhook.ClientConfig.Service != nil {

			svc := crd.Spec.Conversion.Webhook.ClientConfig.Service
			// Get the webhook service
			_, err := a.Client.GetClient().
				CoreV1().
				Services(svc.Namespace).
				Get(a.Context, svc.Name, v1.GetOptions{})
			if err != nil {
				// If the service is not found, can't create the custom resource
				failures = append(failures, common.Failure{
					Text:          fmt.Sprintf(serviceNotFound, svc.Name),
					KubernetesDoc: apiDoc.GetApiDocV2(apiSpecWebhookService),
					Sensitive: []common.Sensitive{
						{
							Unmasked: svc.Namespace,
							Masked:   util.MaskString(svc.Namespace),
						},
						{
							Unmasked: svc.Name,
							Masked:   util.MaskString(svc.Name),
						},
					},
				})

				AnalyzerErrorsMetric.WithLabelValues(
					crd.Spec.Names.Singular,
					svc.Name,
					svc.Namespace,
				).Set(float64(len(failures)))
			}

		}

		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s", crd.Name)] = common.PreAnalysis{
				FailureDetails: failures,
			}

		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}

		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}
