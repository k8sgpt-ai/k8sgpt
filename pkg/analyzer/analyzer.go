package analyzer

import (
	"context"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/hpa"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/ingress"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/pod"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/pvc"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/rs"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/service"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
)

type IAnalyzer interface {
	Analyze() error
	GetResult() []common.Result
}

var AnalyzerMap = map[string]IAnalyzer{
	"Pod":                   &pod.PodAnalyzer{},
	"ReplicaSet":            &rs.ReplicaSetAnalyzer{},
	"PersistentVolumeClaim": &pvc.PvcAnalyzer{},
	"Service":               &service.ServiceAnalyzer{},
	"Ingress":               &ingress.IngressAnalyzer{},
	"HPA":                   &hpa.HPAAnalyzer{},
}

var coreAnalyzerList = []string{"Pod", "ReplicaSet", "PersistentVolumeClaim", "Service", "Ingress"}
var additionalAnalyzerList = []string{"HPA"}

func NewAnalyzer(analyzer string, client *kubernetes.Client, context context.Context, namespace string, aiClient ai.IAI, explain bool) (IAnalyzer, error) {
	analyzerConfig := common.Analyzer{
		AIClient:  aiClient,
		Namespace: namespace,
		Context:   context,
		Client:    client,
		Explain:   explain,
	}

	analyzerConfig.PreAnalysis = make(map[string]common.PreAnalysis)
	return AnalyzerMap[analyzer], nil
}

func ListFilters() ([]string, []string) {
	coreKeys := []string{}
	for _, filter := range coreAnalyzerList {
		coreKeys = append(coreKeys, filter)
	}

	additionalKeys := []string{}
	for _, filter := range coreAnalyzerList {
		coreKeys = append(additionalKeys, filter)
	}
	return coreKeys, additionalKeys
}

func GetAnalyzerList() []string {
	list := []string{}

	list = append(list, coreAnalyzerList...)
	list = append(list, additionalAnalyzerList...)

	list = removeDuplicateStr(list)

	return list
}
func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
