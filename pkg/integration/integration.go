package integration

import (
	"errors"

	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/integration/trivy"
)

type IIntegration interface {
	// Add adds an integration to the cluster
	Deploy(namespace string) error
	// Remove removes an integration from the cluster
	UnDeploy(namespace string) error
	//
	AddAnalyzer() (string, analyzer.IAnalyzer, error)
	// RemoveAnalyzer removes an analyzer from the cluster
	RemoveAnalyzer() error

	IsActivate() bool
}

type Integration struct {
}

var integrations = map[string]IIntegration{
	"trivy": trivy.NewTrivy(),
}

func NewIntegration() *Integration {
	return &Integration{}
}

func (*Integration) List() []string {
	keys := make([]string, 0, len(integrations))
	for k := range integrations {
		keys = append(keys, k)
	}
	return keys
}

func (*Integration) Activate(name string, namespace string) error {
	if _, ok := integrations[name]; !ok {
		return errors.New("integration not found")
	}
	return integrations[name].Deploy(namespace)
}

func (*Integration) Deactivate(name string, namespace string) error {
	if _, ok := integrations[name]; !ok {
		return errors.New("integration not found")
	}
	return integrations[name].UnDeploy(namespace)
}
