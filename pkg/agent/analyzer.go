package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
)

type AnalyzerAgent struct {
	aiClient ai.IAI
	context  context.Context
	language string
	texts    []string
	prompt   string
}

type AnalyzerAgentConfiguration struct {
	Language string
	Texts    []string
}

func (agent *AnalyzerAgent) Configure(agentConfiguration AgentConfiguration) {
	agent.aiClient = agentConfiguration.AiClient
	agent.context = agentConfiguration.Context
	agent.language = agentConfiguration.Analyzer.Language
	agent.texts = agentConfiguration.Analyzer.Texts
	agent.prompt = `Simplify the following Kubernetes error message delimited by triple dashes written in --- %s --- language; --- %s ---.
	Provide the most possible solution in a step by step style in no more than 280 characters. Write the output in the following format:
	Error: {Explain error here}
	Solution: {Step by step solution here}
	`
}

func (agent AnalyzerAgent) Process() (string, error) {
	inputKey := strings.Join(agent.texts, " ")
	prompt := fmt.Sprintf(strings.TrimSpace(agent.prompt), agent.language, inputKey)
	response, err := agent.aiClient.GetCompletion(agent.context, prompt)
	if err != nil {
		return "", err
	}
	return response, nil
}

func (agent AnalyzerAgent) GetName() string {
	return "analyzer"
}
