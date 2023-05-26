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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type IngressAnalyzer struct{}

func (IngressAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "Ingress"
	apiDoc := kubernetes.K8sApiReference{
		Kind: kind,
		ApiVersion: schema.GroupVersion{
			Group:   "networking",
			Version: "v1",
		},
		Discovery: a.Client.Client.Discovery(),
	}

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	list, err := a.Client.GetClient().NetworkingV1().Ingresses(a.Namespace).List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, ing := range list.Items {
		var failures []common.Failure

		// get ingressClassName
		ingressClassName := ing.Spec.IngressClassName
		if ingressClassName == nil {
			ingClassValue := ing.Annotations["kubernetes.io/ingress.class"]
			if ingClassValue == "" {
				doc := apiDoc.GetApiDocV2("spec.ingressClassName")

				failures = append(failures, common.Failure{
					Text:          fmt.Sprintf("Ingress %s/%s does not specify an Ingress class.", ing.Namespace, ing.Name),
					KubernetesDoc: doc,
					Sensitive: []common.Sensitive{
						{
							Unmasked: ing.Namespace,
							Masked:   util.MaskString(ing.Namespace),
						},
						{
							Unmasked: ing.Name,
							Masked:   util.MaskString(ing.Name),
						},
					},
				})
			} else {
				ingressClassName = &ingClassValue
			}
		}

		// check if ingressclass exist
		if ingressClassName != nil {
			_, err := a.Client.GetClient().NetworkingV1().IngressClasses().Get(a.Context, *ingressClassName, metav1.GetOptions{})
			if err != nil {
				doc := apiDoc.GetApiDocV2("spec.ingressClassName")

				failures = append(failures, common.Failure{
					Text:          fmt.Sprintf("Ingress uses the ingress class %s which does not exist.", *ingressClassName),
					KubernetesDoc: doc,
					Sensitive: []common.Sensitive{
						{
							Unmasked: *ingressClassName,
							Masked:   util.MaskString(*ingressClassName),
						},
					},
				})
			}
		}

		// loop over rules
		for _, rule := range ing.Spec.Rules {
			// loop over paths
			for _, path := range rule.HTTP.Paths {
				_, err := a.Client.GetClient().CoreV1().Services(ing.Namespace).Get(a.Context, path.Backend.Service.Name, metav1.GetOptions{})
				if err != nil {
					doc := apiDoc.GetApiDocV2("spec.rules.http.paths.backend.service")

					failures = append(failures, common.Failure{
						Text:          fmt.Sprintf("Ingress uses the service %s/%s which does not exist.", ing.Namespace, path.Backend.Service.Name),
						KubernetesDoc: doc,
						Sensitive: []common.Sensitive{
							{
								Unmasked: ing.Namespace,
								Masked:   util.MaskString(ing.Namespace),
							},
							{
								Unmasked: path.Backend.Service.Name,
								Masked:   util.MaskString(path.Backend.Service.Name),
							},
						},
					})
				}
			}
		}

		for _, tls := range ing.Spec.TLS {
			_, err := a.Client.GetClient().CoreV1().Secrets(ing.Namespace).Get(a.Context, tls.SecretName, metav1.GetOptions{})
			if err != nil {
				doc := apiDoc.GetApiDocV2("spec.tls.secretName")

				failures = append(failures, common.Failure{
					Text:          fmt.Sprintf("Ingress uses the secret %s/%s as a TLS certificate which does not exist.", ing.Namespace, tls.SecretName),
					KubernetesDoc: doc,
					Sensitive: []common.Sensitive{
						{
							Unmasked: ing.Namespace,
							Masked:   util.MaskString(ing.Namespace),
						},
						{
							Unmasked: tls.SecretName,
							Masked:   util.MaskString(tls.SecretName),
						},
					},
				})
			}
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", ing.Namespace, ing.Name)] = common.PreAnalysis{
				Ingress:        ing,
				FailureDetails: failures,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, ing.Name, ing.Namespace).Set(float64(len(failures)))

		}

	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.Ingress.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}
