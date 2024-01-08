package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
)

const (
	default_prompt = `Simplify the following Kubernetes error message delimited by triple dashes written in --- %s --- language; --- %s ---.
	Provide the most possible solution in a step by step style in no more than 280 characters. Write the output in the following format:
	Error: {Explain error here}
	Solution: {Step by step solution here}
	`
	trivy_vuln_prompt = "Explain the following trivy scan result and the detail risk or root cause of the CVE ID, then provide a solution. Response in %s: %s"
	trivy_conf_prompt = "Explain the following trivy scan result and the detail risk or root cause of the security check, then provide a solution."
)

var PromptMap = map[string]string{
	"default":             default_prompt,
	"VulnerabilityReport": trivy_vuln_prompt, // for Trivy integration, the key should match `Result.Kind` in pkg/common/types.go
	"ConfigAuditReport":   trivy_conf_prompt,
}

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
	Prompt   string
}

func (agent *AnalyzerAgent) Configure(agentConfiguration AgentConfiguration) {
	agent.aiClient = agentConfiguration.AiClient
	agent.context = agentConfiguration.Context
	agent.language = agentConfiguration.Analyzer.Language
	agent.texts = agentConfiguration.Analyzer.Texts
	agent.prompt = agentConfiguration.Analyzer.Prompt
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
