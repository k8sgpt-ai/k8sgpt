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

var analyzerMap = map[string]func(ctx context.Context, config *AnalysisConfiguration,
	client *kubernetes.Client, aiClient ai.IAI, analysisResults *[]Analysis) error{
	"Pod":                   AnalyzePod,
	"ReplicaSet":            AnalyzeReplicaSet,
	"PersistentVolumeClaim": AnalyzePersistentVolumeClaim,
	"Service":               AnalyzeEndpoints,
	"Ingress":               AnalyzeIngress,
}

func RunAnalysis(ctx context.Context, filters []string, config *AnalysisConfiguration,
	client *kubernetes.Client,
	aiClient ai.IAI, analysisResults *[]Analysis) error {

	// if there are no filters selected then run all of them
	if len(filters) == 0 {
		for _, analyzer := range analyzerMap {
			if err := analyzer(ctx, config, client, aiClient, analysisResults); err != nil {
				return err
			}
		}
		return nil
	}

	for _, filter := range filters {
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

func ListFilters() []string {
	keys := make([]string, 0, len(analyzerMap))
	for k := range analyzerMap {
		keys = append(keys, k)
	}
	return keys
}
