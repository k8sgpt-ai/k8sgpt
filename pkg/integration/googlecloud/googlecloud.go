package googlecloud

import (
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/spf13/viper"
)

type GoogleCloud struct{}

func (g *GoogleCloud) Deploy(namespace string) error {
	return nil
}

func (g *GoogleCloud) UnDeploy(namespace string) error {
	return nil
}

func (g *GoogleCloud) GetAnalyzerName() []string {
	return []string{
		"GKEClusterNotificationAnalysis",
	}
}

func (g *GoogleCloud) GetNamespace() (string, error) {
	return "", nil
}

func (g *GoogleCloud) OwnsAnalyzer(s string) bool {
	for _, az := range g.GetAnalyzerName() {
		if s == az {
			return true
		}
	}
	return false
}

func (g *GoogleCloud) IsActivate() bool {
	activeFilters := viper.GetStringSlice("active_filters")

	for _, filter := range g.GetAnalyzerName() {
		for _, af := range activeFilters {
			if af == filter {
				return true
			}
		}
	}

	return false
}

func NewGoogleCloud() *GoogleCloud {
	return &GoogleCloud{}
}

func (g *GoogleCloud) AddAnalyzer(mergedMap *map[string]common.IAnalyzer) {
	enableClusterNotificationAnalysis := true
	if viper.Get("googlecloud.enable_gke_clusternotificationanalysis") != nil {
		enableClusterNotificationAnalysis = viper.GetBool("googlecloud.enable_gke_clusternotificationanalysis")
	}
	(*mergedMap)["GKEClusterNotificationAnalysis"] = &GKEAnalyzer{
		EnableClusterNotificationAnalysis: enableClusterNotificationAnalysis,
	}
}
