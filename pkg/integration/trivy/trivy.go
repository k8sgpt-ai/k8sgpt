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

package trivy

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	helmclient "github.com/mittwald/go-helm-client"
	"github.com/spf13/viper"
	"helm.sh/helm/v3/pkg/repo"
)

var (
	Repo          = getEnv("TRIVY_REPO", "https://aquasecurity.github.io/helm-charts/")
	Version       = getEnv("TRIVY_VERSION", "0.13.0")
	ChartName     = getEnv("TRIVY_CHART_NAME", "trivy-operator")
	RepoShortName = getEnv("TRIVY_REPO_SHORT_NAME", "aqua")
	ReleaseName   = getEnv("TRIVY_RELEASE_NAME", "trivy-operator-k8sgpt")
)

type Trivy struct {
	helm helmclient.Client
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func NewTrivy() *Trivy {
	helmClient, err := helmclient.New(&helmclient.Options{})
	if err != nil {
		panic(err)
	}
	return &Trivy{
		helm: helmClient,
	}
}

func (t *Trivy) GetAnalyzerName() []string {
	return []string{
		"VulnerabilityReport",
		"ConfigAuditReport",
	}
}

// This doesnt work
func (t *Trivy) GetNamespace() (string, error) {
	releases, err := t.helm.ListDeployedReleases()
	if err != nil {
		return "", err
	}
	for _, rel := range releases {
		if rel.Name == ReleaseName {
			return rel.Namespace, nil
		}
	}
	return "", status.Error(codes.NotFound, "trivy release not found")
}

func (t *Trivy) OwnsAnalyzer(analyzer string) bool {

	for _, a := range t.GetAnalyzerName() {
		if analyzer == a {
			return true
		}
	}
	return false
}
func (t *Trivy) Deploy(namespace string) error {

	// Add the repository
	chartRepo := repo.Entry{
		Name: RepoShortName,
		URL:  Repo,
	}
	// Add a chart-repository to the client.
	if err := t.helm.AddOrUpdateChartRepo(chartRepo); err != nil {
		panic(err)
	}

	chartSpec := helmclient.ChartSpec{
		ReleaseName: ReleaseName,
		ChartName:   fmt.Sprintf("%s/%s", RepoShortName, ChartName),
		Namespace:   namespace,

		//TODO: All of this should be configurable
		UpgradeCRDs:     true,
		Wait:            false,
		Timeout:         300,
		CreateNamespace: true,
	}

	// Install a chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	if _, err := t.helm.InstallOrUpgradeChart(context.Background(), &chartSpec, nil); err != nil {
		return err
	}

	return nil
}

func (t *Trivy) UnDeploy(namespace string) error {
	chartSpec := helmclient.ChartSpec{
		ReleaseName: ReleaseName,
		ChartName:   fmt.Sprintf("%s/%s", RepoShortName, ChartName),
		Namespace:   namespace,
		UpgradeCRDs: true,
		Wait:        false,
		Timeout:     300,
	}
	// Uninstall the chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	if err := t.helm.UninstallRelease(&chartSpec); err != nil {
		return err
	}
	return nil
}

func (t *Trivy) isDeployed() bool {
	// check if aquasec apigroup is available as a marker if trivy is installed on the cluster
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
		if group.Name == "aquasecurity.github.io" {
			return true
		}
	}

	return false
}

func (t *Trivy) isFilterActive() bool {
	activeFilters := viper.GetStringSlice("active_filters")

	for _, filter := range t.GetAnalyzerName() {
		for _, af := range activeFilters {
			if af == filter {
				return true
			}
		}
	}

	return false
}

func (t *Trivy) IsActivate() bool {
	if t.isFilterActive() && t.isDeployed() {
		return true
	} else {
		return false
	}
}

func (t *Trivy) AddAnalyzer(mergedMap *map[string]common.IAnalyzer) {

	(*mergedMap)["VulnerabilityReport"] = &TrivyAnalyzer{
		vulernabilityReportAnalysis: true,
	}
	(*mergedMap)["ConfigAuditReport"] = &TrivyAnalyzer{
		configAuditReportAnalysis: true,
	}

}
