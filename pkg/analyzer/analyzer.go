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

func RunAnalysis(ctx context.Context, client *kubernetes.Client,
	aiClient ai.IAI, explain bool, analysisResults *[]Analysis) error {

	err := AnalyzePod(ctx, client, aiClient, explain, analysisResults)
	if err != nil {
		return err
	}

	err = AnalyzeReplicaSet(ctx, client, aiClient, explain, analysisResults)
	if err != nil {
		return err
	}

	err = AnalyzePersistentVolumeClaim(ctx, client, aiClient, explain, analysisResults)
	if err != nil {
		return err
	}

	err = AnalyzeEndpoints(ctx, client, aiClient, explain, analysisResults)
	if err != nil {
		return err
	}
	return nil
}

func ParseViaAI(ctx context.Context, aiClient ai.IAI, prompt []string,
	nocache bool) (string, error) {
	// parse the text with the AI backend
	inputKey := strings.Join(prompt, " ")
	// Check for cached data
	sEnc := base64.StdEncoding.EncodeToString([]byte(inputKey))
	// find in viper cache
	if viper.IsSet(sEnc) && !nocache {
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
