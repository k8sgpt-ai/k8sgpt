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
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	gtwapi "sigs.k8s.io/gateway-api/apis/v1"
)

type HTTPRouteAnalyzer struct{}

// Gateway analyser will analyse all different Kinds and search for missing object dependencies
func (HTTPRouteAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "HTTPRoute"
	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	routeList := &gtwapi.HTTPRouteList{}
	gtw := &gtwapi.Gateway{}
	service := &corev1.Service{}
	client := a.Client.CtrlClient
	err := gtwapi.AddToScheme(client.Scheme())
	if err != nil {
		return nil, err
	}
	if err := client.List(a.Context, routeList, &ctrl.ListOptions{}); err != nil {
		return nil, err
	}
	var preAnalysis = map[string]common.PreAnalysis{}

	// Find all unhealthy gateway Classes
	for _, route := range routeList.Items {
		var failures []common.Failure

		// Check if Gateways exists in the same or designated namespace
		// TODO: when meshes and ClusterIp options are adopted we can add more checks
		// e.g Service Port matching
		for _, gtwref := range route.Spec.ParentRefs {
			namespace := route.Namespace
			if gtwref.Namespace != nil {
				namespace = string(*gtwref.Namespace)
			}
			err := client.Get(a.Context, ctrl.ObjectKey{Namespace: namespace, Name: string(gtwref.Name)}, gtw, &ctrl.GetOptions{})
			if errors.IsNotFound(err) {
				failures = append(failures, common.Failure{
					Text: fmt.Sprintf(
						"HTTPRoute uses the Gateway '%s/%s' which does not exist in the same namespace.",
						namespace,
						gtwref.Name,
					),
					Sensitive: []common.Sensitive{
						{
							Unmasked: gtw.Namespace,
							Masked:   util.MaskString(gtw.Namespace),
						},
						{
							Unmasked: gtw.Name,
							Masked:   util.MaskString(gtw.Name),
						},
					},
				})
			} else {
				// Check if the aforementioned Gateway allows the HTTPRoutes from the route's namespace
				for _, listener := range gtw.Spec.Listeners {
					if listener.AllowedRoutes.Namespaces != nil {
						switch allow := listener.AllowedRoutes.Namespaces.From; {
						case *allow == gtwapi.NamespacesFromSame:
							// check if Gateway is in the same namespace
							if route.Namespace != gtw.Namespace {
								failures = append(failures, common.Failure{
									Text: fmt.Sprintf("HTTPRoute '%s/%s' is deployed in a different namespace from Gateway '%s/%s' which only allows HTTPRoutes from its namespace.",
										route.Namespace,
										route.Name,
										gtw.Namespace,
										gtw.Name,
									),
									Sensitive: []common.Sensitive{
										{
											Unmasked: route.Namespace,
											Masked:   util.MaskString(route.Namespace),
										},
										{
											Unmasked: route.Name,
											Masked:   util.MaskString(route.Name),
										},
										{
											Unmasked: gtw.Namespace,
											Masked:   util.MaskString(gtw.Namespace),
										},
										{
											Unmasked: gtw.Name,
											Masked:   util.MaskString(gtw.Name),
										},
									},
								})
							}
						case *allow == gtwapi.NamespacesFromSelector:
							// check if our route include the same selector Label
							if !util.LabelsIncludeAny(listener.AllowedRoutes.Namespaces.Selector.MatchLabels, route.Labels) {
								failures = append(failures, common.Failure{
									Text: fmt.Sprintf(
										"HTTPRoute '%s/%s' can't be attached on Gateway '%s/%s', selector labels do not match HTTProute's labels.",
										route.Namespace,
										route.Name,
										gtw.Namespace,
										gtw.Name,
									),
									Sensitive: []common.Sensitive{
										{
											Unmasked: route.Namespace,
											Masked:   util.MaskString(route.Namespace),
										},
										{
											Unmasked: route.Name,
											Masked:   util.MaskString(route.Name),
										},
										{
											Unmasked: gtw.Namespace,
											Masked:   util.MaskString(gtw.Namespace),
										},
										{
											Unmasked: gtw.Name,
											Masked:   util.MaskString(gtw.Name),
										},
									},
								})

							}

						}
					}
				}
			}

		}
		// Check if the Backends are valid services and ports are matching with services Ports
		for _, rule := range route.Spec.Rules {
			for _, backend := range rule.BackendRefs {
				err := client.Get(a.Context, ctrl.ObjectKey{Namespace: route.Namespace, Name: string(backend.Name)}, service, &ctrl.GetOptions{})
				if errors.IsNotFound(err) {
					failures = append(failures, common.Failure{
						Text: fmt.Sprintf(
							"HTTPRoute uses the Service '%s/%s' which does not exist.",
							route.Namespace,
							backend.Name,
						),
						Sensitive: []common.Sensitive{
							{
								Unmasked: service.Namespace,
								Masked:   util.MaskString(service.Namespace),
							},
							{
								Unmasked: service.Name,
								Masked:   util.MaskString(service.Name),
							},
						},
					})
				} else {
					portMatch := false
					for _, svcPort := range service.Spec.Ports {
						if int32(*backend.Port) == svcPort.Port {
							portMatch = true
						}
					}
					if !portMatch {
						failures = append(failures, common.Failure{
							Text: fmt.Sprintf(
								"HTTPRoute's backend service '%s' is using port '%d' but the corresponding K8s service '%s/%s' isn't configured with the same port.",
								backend.Name,
								int32(*backend.Port),
								service.Namespace,
								service.Name,
							),
							Sensitive: []common.Sensitive{
								{
									Unmasked: string(backend.Name),
									Masked:   util.MaskString(string(backend.Name)),
								},
								{
									Unmasked: service.Name,
									Masked:   util.MaskString(service.Name),
								},
								{
									Unmasked: service.Namespace,
									Masked:   service.Namespace,
								},
							},
						})
					}
				}
			}
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", route.Namespace, route.Name)] = common.PreAnalysis{
				HTTPRoute:      route,
				FailureDetails: failures,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, route.Name, route.Namespace).Set(float64(len(failures)))
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
