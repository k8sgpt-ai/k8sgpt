package server

import (
	"context"
	json "encoding/json"

	schemav1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analysis"
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

	config, err := analysis.NewAnalysis(
		i.Backend,
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
