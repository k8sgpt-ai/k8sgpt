package analyzer

import (
	"context"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
)

type IAnalyzer interface {
	RunAnalysis(ctx context.Context, config *AnalysisConfiguration, client *kubernetes.Client, aiClient ai.IAI,
		analysisResults *[]Analysis) error
}
