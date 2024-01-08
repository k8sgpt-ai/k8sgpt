package agent

import (
	"context"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
)

type AgentConfiguration struct {
	AiClient  ai.IAI
	Context   context.Context
	Router    RouterAgentConfiguration
	Analyzer  AnalyzerAgentConfiguration
	Evaluator EvaluatorAgentConfiguration
}
