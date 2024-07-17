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

package kyverno

import (
	"os"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/spf13/viper"
)

type Kyverno struct{}

func NewKyverno() *Kyverno {
	return &Kyverno{}
}

func (k *Kyverno) GetAnalyzerName() []string {
	return []string{
		//from wgpolicyk8s.io/v1alpha2
		"PolicyReport",
		"ClusterPolicyReport",
	}
}

func (k *Kyverno) OwnsAnalyzer(analyzer string) bool {

	for _, a := range k.GetAnalyzerName() {
		if analyzer == a {
			return true
		}
	}
	return false
}

func (k *Kyverno) isDeployed() bool {
	// check if wgpolicyk8s apigroup is available as a marker if new policy resource available is installed on the cluster
	kubecontext := viper.GetString("kubecontext")
	kubeconfig := viper.GetString("kubeconfig")
	client, err := kubernetes.NewClient(kubecontext, kubeconfig)
	if err != nil {
		// TODO: better error handling
		color.Red("Error initialising kubernetes client: %v", err)
		os.Exit(1)
	}
	groups, _, err := client.Client.Discovery().ServerGroupsAndResources()
	if err != nil {
		// TODO: better error handling
		color.Red("Error initialising discovery client: %v", err)
		os.Exit(1)
	}

	for _, group := range groups {
		if group.Name == "kyverno.io" {
			return true
		}
	}

	return false
}

func (k *Kyverno) isFilterActive() bool {
	activeFilters := viper.GetStringSlice("active_filters")

	for _, filter := range k.GetAnalyzerName() {
		for _, af := range activeFilters {
			if af == filter {
				return true
			}
		}
	}

	return false
}

func (k *Kyverno) IsActivate() bool {
	if k.isFilterActive() && k.isDeployed() {
		return true
	} else {
		return false
	}
}

func (k *Kyverno) AddAnalyzer(mergedMap *map[string]common.IAnalyzer) {

	(*mergedMap)["PolicyReport"] = &KyvernoAnalyzer{
		policyReportAnalysis: true,
	}
	(*mergedMap)["ClusterPolicyReport"] = &KyvernoAnalyzer{
		clusterReportAnalysis: true,
	}
}

func (k *Kyverno) Deploy(namespace string) error {
	return nil
}

func (k *Kyverno) UnDeploy(_ string) error {
	return nil
}

func (t *Kyverno) GetNamespace() (string, error) {
	return "", nil
}
