package analyzer

type IAnalyzer interface {
	Analyze(analysis Analyzer) ([]Result, error)
}

var coreAnalyzerMap = map[string]IAnalyzer{
	"Pod":                   PodAnalyzer{},
	"ReplicaSet":            ReplicaSetAnalyzer{},
	"PersistentVolumeClaim": PvcAnalyzer{},
	"Service":               ServiceAnalyzer{},
	"Ingress":               IngressAnalyzer{},
}

var additionalAnalyzerMap = map[string]IAnalyzer{
	"HorizontalPodAutoScaler": HpaAnalyzer{},
	"PodDisruptionBudget":     PdbAnalyzer{},
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
