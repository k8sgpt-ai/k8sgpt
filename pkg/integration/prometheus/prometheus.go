package prometheus

import (
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/spf13/viper"
)

const (
	ConfigValidate = "PrometheusConfigValidate"
	ConfigRelabel  = "PrometheusConfigRelabelReport"
)

type Prometheus struct {
}

func NewPrometheus() *Prometheus {
	return &Prometheus{}
}

func (p *Prometheus) Deploy(_ string) error {
	// no-op
	// We just care about existing deployments.
	return nil
}

func (p *Prometheus) UnDeploy(_ string) error {
	// no-op
	// We just care about existing deployments.
	return nil
}

func (p *Prometheus) AddAnalyzer(mergedMap *map[string]common.IAnalyzer) {
	(*mergedMap)[ConfigValidate] = &ConfigAnalyzer{}
	(*mergedMap)[ConfigRelabel] = &RelabelAnalyzer{}
}

func (p *Prometheus) GetAnalyzerName() []string {
	return []string{ConfigValidate, ConfigRelabel}
}

func (p *Prometheus) GetNamespace() (string, error) {
	return "", nil
}

func (p *Prometheus) OwnsAnalyzer(analyzer string) bool {
	return (analyzer == ConfigValidate) || (analyzer == ConfigRelabel)
}

func (t *Prometheus) IsActivate() bool {
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
