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

package integration

import (
	"errors"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/integration/trivy"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	"github.com/spf13/viper"
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

	GetAnalyzerName() string

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

func (*Integration) Activate(name string, namespace string, activeFilters []string) error {
	if _, ok := integrations[name]; !ok {
		return errors.New("integration not found")
	}

	mergedFilters := append(activeFilters, integrations[name].GetAnalyzerName())

	uniqueFilters, dupplicatedFilters := util.RemoveDuplicates(mergedFilters)

	// Verify dupplicate
	if len(dupplicatedFilters) != 0 {
		color.Red("Integration already activated : %s", strings.Join(dupplicatedFilters, ", "))
		os.Exit(1)
	}

	viper.Set("active_filters", uniqueFilters)

	if err := integrations[name].Deploy(namespace); err != nil {
		return err
	}

	if err := viper.WriteConfig(); err != nil {
		color.Red("Error writing config file: %s", err.Error())
		os.Exit(1)
	}

	return nil
}

func (*Integration) Deactivate(name string, namespace string) error {
	if _, ok := integrations[name]; !ok {
		return errors.New("integration not found")
	}

	activeFilters := viper.GetStringSlice("active_filters")

	// Update filters
	// This might be a bad idea, but we cannot reference analyzer here
	foundFilter := false
	for i, v := range activeFilters {
		if v == integrations[name].GetAnalyzerName() {
			foundFilter = true
			activeFilters = append(activeFilters[:i], activeFilters[i+1:]...)
			break
		}
	}
	if !foundFilter {
		color.Red("Ingregation %s does not exist in configuration file. Please use k8sgpt integration add.", name)
		os.Exit(1)
	}

	if err := integrations[name].UnDeploy(namespace); err != nil {
		return err
	}

	viper.Set("active_filters", activeFilters)

	if err := viper.WriteConfig(); err != nil {
		color.Red("Error writing config file: %s", err.Error())
		os.Exit(1)
	}

	return nil
}

func (*Integration) IsActivate(name string) (bool, error) {
	if _, ok := integrations[name]; !ok {
		return false, errors.New("integration not found")
	}
	return integrations[name].IsActivate(), nil
}
