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
	"os"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/integration"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	AnalyzerErrorsMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "analyzer_errors",
		Help: "Number of errors detected by analyzer",
	}, []string{"analyzer_name", "object_name", "namespace"})
)

var coreAnalyzerMap = map[string]common.IAnalyzer{
	"Pod":                            PodAnalyzer{},
	"Deployment":                     DeploymentAnalyzer{},
	"ReplicaSet":                     ReplicaSetAnalyzer{},
	"PersistentVolumeClaim":          PvcAnalyzer{},
	"Service":                        ServiceAnalyzer{},
	"Ingress":                        IngressAnalyzer{},
	"StatefulSet":                    StatefulSetAnalyzer{},
	"CronJob":                        CronJobAnalyzer{},
	"Node":                           NodeAnalyzer{},
	"ValidatingWebhookConfiguration": ValidatingWebhookAnalyzer{},
	"MutatingWebhookConfiguration":   MutatingWebhookAnalyzer{},
}

var additionalAnalyzerMap = map[string]common.IAnalyzer{
	"HorizontalPodAutoScaler": HpaAnalyzer{},
	"PodDisruptionBudget":     PdbAnalyzer{},
	"NetworkPolicy":           NetworkPolicyAnalyzer{},
	"Log":                     LogAnalyzer{},
	"GatewayClass":            GatewayClassAnalyzer{},
	"Gateway":                 GatewayAnalyzer{},
	"HTTPRoute":               HTTPRouteAnalyzer{},
}

func ListFilters() ([]string, []string, []string) {
	coreKeys := make([]string, 0, len(coreAnalyzerMap))
	for k := range coreAnalyzerMap {
		coreKeys = append(coreKeys, k)
	}

	additionalKeys := make([]string, 0, len(additionalAnalyzerMap))
	for k := range additionalAnalyzerMap {
		additionalKeys = append(additionalKeys, k)
	}

	integrationProvider := integration.NewIntegration()
	var integrationAnalyzers []string

	for _, i := range integrationProvider.List() {
		b, _ := integrationProvider.IsActivate(i)
		if b {
			in, err := integrationProvider.Get(i)
			if err != nil {
				fmt.Println(color.RedString(err.Error()))
				os.Exit(1)
			}
			integrationAnalyzers = append(integrationAnalyzers, in.GetAnalyzerName()...)
		}
	}

	return coreKeys, additionalKeys, integrationAnalyzers
}

func GetAnalyzerMap() (map[string]common.IAnalyzer, map[string]common.IAnalyzer) {

	coreAnalyzer := make(map[string]common.IAnalyzer)
	mergedAnalyzerMap := make(map[string]common.IAnalyzer)

	// add core analyzer
	for key, value := range coreAnalyzerMap {
		coreAnalyzer[key] = value
		mergedAnalyzerMap[key] = value
	}

	// add additional analyzer
	for key, value := range additionalAnalyzerMap {
		mergedAnalyzerMap[key] = value
	}

	integrationProvider := integration.NewIntegration()

	for _, i := range integrationProvider.List() {
		b, err := integrationProvider.IsActivate(i)
		if err != nil {
			fmt.Println(color.RedString(err.Error()))
			os.Exit(1)
		}
		if b {
			in, err := integrationProvider.Get(i)
			if err != nil {
				fmt.Println(color.RedString(err.Error()))
				os.Exit(1)
			}
			in.AddAnalyzer(&mergedAnalyzerMap)
		}
	}

	return coreAnalyzer, mergedAnalyzerMap
}
