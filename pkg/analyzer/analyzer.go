package analyzer

import (
	"context"
	"fmt"
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

const (
	PodAnalyzerName                   = "Pod"
	ReplicaSetAnalyzerName            = "ReplicaSet"
	PersistentVolumeClaimAnalyzerName = "PersistentVolumeClaim"
	ServiceAnalyzerName               = "Service"
	IngressAnalyzerName               = "Ingress"
	HPAAnalyzerName                   = "HorizontalPodAutoScaler"
)

var (
	coreAnalyzerList = []string{
		PodAnalyzerName,
		ReplicaSetAnalyzerName,
		PersistentVolumeClaimAnalyzerName,
		ServiceAnalyzerName,
		IngressAnalyzerName,
		HPAAnalyzerName,
	}

	additionalAnalyzers = []string{
		HPAAnalyzerName,
	}
)

func NewAnalyzer(analyzer string, client *kubernetes.Client, context context.Context, namespace string) (IAnalyzer, error) {
	analyzerConfig := common.Analyzer{
		Namespace: namespace,
		Context:   context,
		Client:    client,
	}

	analyzerConfig.PreAnalysis = make(map[string]common.PreAnalysis)

	switch analyzer {
	case PodAnalyzerName:
		return &pod.PodAnalyzer{
			Analyzer: analyzerConfig,
		}, nil
	case ReplicaSetAnalyzerName:
		return &rs.ReplicaSetAnalyzer{
			Analyzer: analyzerConfig,
		}, nil
	case IngressAnalyzerName:
		return &ingress.IngressAnalyzer{
			Analyzer: analyzerConfig,
		}, nil
	case HPAAnalyzerName:
		return &hpa.HPAAnalyzer{
			Analyzer: analyzerConfig,
		}, nil
	case PersistentVolumeClaimAnalyzerName:
		return &pvc.PvcAnalyzer{
			Analyzer: analyzerConfig,
		}, nil
	case ServiceAnalyzerName:
		return &service.ServiceAnalyzer{
			Analyzer: analyzerConfig,
		}, nil
	default:
		return nil, fmt.Errorf("Analyzer %s not supported", analyzer)
	}
}

/*
func ParseViaAI(ctx context.Context, config *analysis.Analysis,

		aiClient ai.IAI, prompt []string) (string, error) {
		// parse the text with the AI backend
		inputKey := strings.Join(prompt, " ")
		// Check for cached data
		sEnc := base64.StdEncoding.EncodeToString([]byte(inputKey))
		// find in viper cache
		if viper.IsSet(sEnc) && !config.NoCache {
			// retrieve data from cache
			response := viper.GetString(sEnc)
			if response == "" {
				color.Red("error retrieving cached data")
				return "", nil
			}
			output, err := base64.StdEncoding.DecodeString(response)
			if err != nil {
				color.Red("error decoding cached data: %v", err)
				return "", nil
			}
			return string(output), nil
		}

		response, err := aiClient.GetCompletion(ctx, inputKey)
		if err != nil {
			color.Red("error getting completion: %v", err)
			return "", err
		}

		if !viper.IsSet(sEnc) {
			viper.Set(sEnc, base64.StdEncoding.EncodeToString([]byte(response)))
			if err := viper.WriteConfig(); err != nil {
				color.Red("error writing config: %v", err)
				return "", nil
			}
		}
		return response, nil
	}
*/
func ListFilters() ([]string, []string) {
	coreKeys := []string{}
	for _, filter := range coreAnalyzerList {
		coreKeys = append(coreKeys, filter)
	}

	additionalKeys := []string{}
	for _, filter := range additionalAnalyzers {
		coreKeys = append(additionalKeys, filter)
	}
	return coreKeys, additionalKeys
}

func GetAnalyzerList() []string {
	list := []string{}

	list = append(list, coreAnalyzerList...)
	list = append(list, additionalAnalyzers...)

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
