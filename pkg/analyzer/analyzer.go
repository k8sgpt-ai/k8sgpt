package analyzer

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/spf13/viper"
)

var coreAnalyzerMap = map[string]func(ctx context.Context, config *AnalysisConfiguration,
	client *kubernetes.Client, aiClient ai.IAI, analysisResults *[]Analysis) error{
	"Pod":                   AnalyzePod,
	"ReplicaSet":            AnalyzeReplicaSet,
	"PersistentVolumeClaim": AnalyzePersistentVolumeClaim,
	"Service":               AnalyzeEndpoints,
	"Ingress":               AnalyzeIngress,
}

var additionalAnalyzerMap = map[string]func(ctx context.Context, config *AnalysisConfiguration,
	client *kubernetes.Client, aiClient ai.IAI, analysisResults *[]Analysis) error{
	"HorizontalPodAutoScaler": AnalyzeHpa,
}

func RunAnalysis(ctx context.Context, filters []string, config *AnalysisConfiguration,
	client *kubernetes.Client,
	aiClient ai.IAI, analysisResults *[]Analysis) error {

	activeFilters := viper.GetStringSlice("active_filters")

	analyzerMap := getAnalyzerMap()

	// if there are no filters selected and no active_filters then run all of them
	if len(filters) == 0 && len(activeFilters) == 0 {
		for _, analyzer := range analyzerMap {
			if err := analyzer(ctx, config, client, aiClient, analysisResults); err != nil {
				return err
			}
		}
		return nil
	}

	// if the filters flag is specified
	if len(filters) != 0 {
		for _, filter := range filters {
			if analyzer, ok := analyzerMap[filter]; ok {
				if err := analyzer(ctx, config, client, aiClient, analysisResults); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// use active_filters
	for _, filter := range activeFilters {
		if analyzer, ok := analyzerMap[filter]; ok {
			if err := analyzer(ctx, config, client, aiClient, analysisResults); err != nil {
				return err
			}
		}
	}
	return nil
}

func ParseViaAI(ctx context.Context, config *AnalysisConfiguration,
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

func getAnalyzerMap() map[string]func(ctx context.Context, config *AnalysisConfiguration,
	client *kubernetes.Client, aiClient ai.IAI, analysisResults *[]Analysis) error {

	mergedMap := make(map[string]func(ctx context.Context, config *AnalysisConfiguration,
		client *kubernetes.Client, aiClient ai.IAI, analysisResults *[]Analysis) error)

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
