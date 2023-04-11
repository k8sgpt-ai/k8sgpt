package trivy

import (
	"context"
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	helmclient "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/repo"
)

const (
	Repo          = "https://aquasecurity.github.io/helm-charts/"
	Version       = "0.13.0"
	ChartName     = "trivy-operator"
	RepoShortName = "aqua"
	ReleaseName   = "trivy-operator-k8sgpt"
)

type Trivy struct {
	helm helmclient.Client
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

func (t *Trivy) GetAnalyzerName() string {
	return "VulnerabilityReport"
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
		UpgradeCRDs: true,
		Wait:        false,
		Timeout:     300,
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

func (t *Trivy) IsActivate() bool {

	if _, err := t.helm.GetRelease(ReleaseName); err != nil {
		return false
	}

	return true
}

func (t *Trivy) AddAnalyzer(mergedMap *map[string]common.IAnalyzer) {

	(*mergedMap)["VulnerabilityReport"] = &TrivyAnalyzer{}

}

func (t *Trivy) RemoveAnalyzer() error {
	return nil
}
