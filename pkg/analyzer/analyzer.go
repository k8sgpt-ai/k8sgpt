package analyzer

import (
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/hpa"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/ingress"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/pdb"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/pod"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/pvc"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/rs"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/service"
)

type IAnalyzer interface {
	Analyze(analysis common.Analyzer) ([]common.Result, error)
}

var coreAnalyzerMap = map[string]IAnalyzer{
	"Pod":                   pod.PodAnalyzer{},
	"ReplicaSet":            rs.ReplicaSetAnalyzer{},
	"PersistentVolumeClaim": pvc.PvcAnalyzer{},
	"Service":               service.ServiceAnalyzer{},
	"Ingress":               ingress.IngressAnalyzer{},
}

var additionalAnalyzerMap = map[string]IAnalyzer{
	"HorizontalPodAutoScaler": hpa.HpaAnalyzer{},
	"PodDisruptionBudget":     pdb.PdbAnalyzer{},
}

func ListFilters() ([]string, []string) {
	coreKeys := make([]string, 0, len(coreAnalyzerMap))
	for k := range coreAnalyzerMap {
		coreKeys = append(coreKeys, k)
	}

	additionalKeys := make([]string, 0, len(additionalAnalyzerMap))
	for k := range additionalAnalyzerMap {
		additionalKeys = append(additionalKeys, k)
	}
	return coreKeys, additionalKeys
}

func GetAnalyzerMap() map[string]IAnalyzer {

	mergedMap := make(map[string]IAnalyzer)

	// add core analyzer
	for key, value := range coreAnalyzerMap {
		mergedMap[key] = value
	}

	// add additional analyzer
	for key, value := range additionalAnalyzerMap {
		mergedMap[key] = value
	}

	return mergedMap
}
