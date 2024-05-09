package server

import (
	"context"
	json "encoding/json"

	schemav1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analysis"
	"github.com/spf13/viper"
)

func (h *handler) Analyze(ctx context.Context, i *schemav1.AnalyzeRequest) (
	*schemav1.AnalyzeResponse,
	error,
) {
	if i.Output == "" {
		i.Output = "json"
	}

	if int(i.MaxConcurrency) == 0 {
		i.MaxConcurrency = 10
	}

	var configAI ai.AIConfiguration
	if err := viper.UnmarshalKey("ai", &configAI); err != nil {
		return nil, err
	}

	// TODO: Include the "ConfigName" field in the AnalyzeRequest data structure
	configName := "default"
	for _, provider := range configAI.Providers {
		if i.Backend == provider.Backend {
			configName = provider.Configs[provider.DefaultConfig].Name
		}
	}

	config, err := analysis.NewAnalysis(
		i.Backend,
		configName,
		i.Language,
		i.Filters,
		i.Namespace,
		i.Nocache,
		i.Explain,
		int(i.MaxConcurrency),
		false, // Kubernetes Doc disabled in server mode
		false, // Interactive mode disabled in server mode
	)
	config.Context = ctx // Replace context for correct timeouts.
	if err != nil {
		return &schemav1.AnalyzeResponse{}, err
	}
	defer config.Close()

	config.RunAnalysis()

	if i.Explain {
		err := config.GetAIResults(i.Output, i.Anonymize)
		if err != nil {
			return &schemav1.AnalyzeResponse{}, err
		}
	}

	out, err := config.PrintOutput(i.Output)
	if err != nil {
		return &schemav1.AnalyzeResponse{}, err
	}
	var obj schemav1.AnalyzeResponse

	err = json.Unmarshal(out, &obj)
	if err != nil {
		return &schemav1.AnalyzeResponse{}, err
	}

	return &obj, nil
}
