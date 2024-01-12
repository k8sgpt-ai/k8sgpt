package prometheus

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
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

func (p *Prometheus) Deploy(namespace string) error {
	// no-op
	color.Green("Activating prometheus integration...")
	// TODO(pintohutch): add timeout or inherit an upstream context
	// for better signal management.
	ctx := context.Background()
	kubecontext := viper.GetString("kubecontext")
	kubeconfig := viper.GetString("kubeconfig")
	client, err := kubernetes.NewClient(kubecontext, kubeconfig)
	if err != nil {
		color.Red("Error initialising kubernetes client: %v", err)
		os.Exit(1)
	}

	// We just care about existing deployments.
	// Try and find Prometheus configurations in the cluster using the provided namespace.
	//
	// Note: We could cache this state and inject it into the various analyzers
	// to save additional parsing later.
	// However, the state of the cluster can change from activation to analysis,
	// so we would want to run this again on each analyze call anyway.
	//
	// One consequence of this is one can run `activate` in one namespace
	// and run `analyze` in another, without issues, as long as Prometheus
	// is found in both.
	// We accept this as a trade-off for the time-being to avoid having the tool
	// manage Prometheus on the behalf of users.
	podConfigs, err := findPrometheusPodConfigs(ctx, client.GetClient(), namespace)
	if err != nil {
		color.Red("Error discovering Prometheus worklads: %v", err)
		os.Exit(1)
	}
	if len(podConfigs) == 0 {
		color.Yellow(fmt.Sprintf(`Prometheus installation not found in namespace: %s.
		Please ensure Prometheus is deployed to analyze.`, namespace))
		return errors.New("no prometheus installation found")
	}
	// Prime state of the analyzer so
	color.Green("Found existing installation")
	return nil
}

func (p *Prometheus) UnDeploy(_ string) error {
	// no-op
	// We just care about existing deployments.
	color.Yellow("Integration will leave Prometheus resources deployed. This is an effective no-op in the cluster.")
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
