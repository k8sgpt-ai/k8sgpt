package analyzer

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/integration"
)

var coreAnalyzerMap = map[string]common.IAnalyzer{
	"Pod":                   PodAnalyzer{},
	"ReplicaSet":            ReplicaSetAnalyzer{},
	"PersistentVolumeClaim": PvcAnalyzer{},
	"Service":               ServiceAnalyzer{},
	"Ingress":               IngressAnalyzer{},
	"StatefulSet":           StatefulSetAnalyzer{},
}

var additionalAnalyzerMap = map[string]common.IAnalyzer{
	"HorizontalPodAutoScaler": HpaAnalyzer{},
	"PodDisruptionBudget":     PdbAnalyzer{},
}

func ListFilters() ([]string, []string, []string) {
	coreKeys := make([]string, 0, len(coreAnalyzerMap))
	for k := range coreAnalyzerMap {
		coreKeys = append(coreKeys, k)
	}

	additionalKeys := make([]string, 0, len(additionalAnalyzerMap))
	for k := range additionalAnalyzerMap {
		additionalKeys = append(additionalKeys, k)
	}

	integrationProvider := integration.NewIntegration()
	var integrationAnalyzers []string

	for _, i := range integrationProvider.List() {
		b, _ := integrationProvider.IsActivate(i)
		if b {
			in, err := integrationProvider.Get(i)
			if err != nil {
				fmt.Println(color.RedString(err.Error()))
				os.Exit(1)
			}
			integrationAnalyzers = append(integrationAnalyzers, in.GetAnalyzerName())
		}
	}

	return coreKeys, additionalKeys, integrationAnalyzers
}

func GetAnalyzerMap() map[string]common.IAnalyzer {

	mergedMap := make(map[string]common.IAnalyzer)

	// add core analyzer
	for key, value := range coreAnalyzerMap {
		mergedMap[key] = value
	}

	// add additional analyzer
	for key, value := range additionalAnalyzerMap {
		mergedMap[key] = value
	}

	integrationProvider := integration.NewIntegration()

	for _, i := range integrationProvider.List() {
		b, err := integrationProvider.IsActivate(i)
		if err != nil {
			fmt.Println(color.RedString(err.Error()))
			os.Exit(1)
		}
		if b {
			in, err := integrationProvider.Get(i)
			if err != nil {
				fmt.Println(color.RedString(err.Error()))
				os.Exit(1)
			}
			in.AddAnalyzer(&mergedMap)
		}
	}

	return mergedMap
}
