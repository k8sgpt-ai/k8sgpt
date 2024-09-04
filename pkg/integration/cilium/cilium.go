package cilium

import (
	"os"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/spf13/viper"
)

const (
	CiliumAnalyzerName string = "CiliumStatus"
)

type Cilium struct{}

func NewCilium() *Cilium {
	return &Cilium{}
}

func (c *Cilium) GetAnalyzerName() []string {
	return []string{
		CiliumAnalyzerName,
	}
}

func (c *Cilium) AddAnalyzer(mergedMap *map[string]common.IAnalyzer) {
	(*mergedMap)[CiliumAnalyzerName] = &CiliumAnalyzer{}
}

func (c *Cilium) OwnsAnalyzer(analyzer string) bool {
	for _, a := range c.GetAnalyzerName() {
		if analyzer == a {
			return true
		}
	}
	return false
}

func (c *Cilium) isDeployed() bool {
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
		if group.Name == "cilium.io" {
			return true
		}
	}

	return false
}

func (c *Cilium) isFilterActive() bool {
	activeFilters := viper.GetStringSlice("active_filters")
	for _, filter := range c.GetAnalyzerName() {
		for _, af := range activeFilters {
			if af == filter {
				return true
			}
		}
	}

	return false
}

func (c *Cilium) IsActivate() bool {
	if c.isFilterActive() && c.isDeployed() {
		return true
	} else {
		return false
	}
}

func (c *Cilium) Deploy(namespace string) error {
	return nil
}

func (c *Cilium) UnDeploy(_ string) error {
	return nil
}

func (c *Cilium) GetNamespace() (string, error) {
	return "", nil
}
