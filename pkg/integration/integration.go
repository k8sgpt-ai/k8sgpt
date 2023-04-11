package integration

import (
	"errors"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/integration/trivy"
)

type IIntegration interface {
	// Add adds an integration to the cluster
	Deploy(namespace string) error
	// Remove removes an integration from the cluster
	UnDeploy(namespace string) error
	//
	AddAnalyzer(*map[string]common.IAnalyzer)
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

func (*Integration) Get(name string) (IIntegration, error) {
	if _, ok := integrations[name]; !ok {
		return nil, errors.New("integration not found")
	}
	return integrations[name], nil
}

func (*Integration) Activate(name string, namespace string) error {
	if _, ok := integrations[name]; !ok {
		return errors.New("integration not found")
	}

	if err := integrations[name].Deploy(namespace); err != nil {
		return err
	}

	return nil
}

func (*Integration) Deactivate(name string, namespace string) error {
	if _, ok := integrations[name]; !ok {
		return errors.New("integration not found")
	}

	if err := integrations[name].UnDeploy(namespace); err != nil {
		return err
	}

	return nil
}

func (*Integration) IsActivate(name string) (bool, error) {
	if _, ok := integrations[name]; !ok {
		return false, errors.New("integration not found")
	}
	return integrations[name].IsActivate(), nil
}
