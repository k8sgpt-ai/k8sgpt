package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
)

type EvaluatorAgent struct {
	aiClient ai.IAI
	context  context.Context
	alert    string
	result   string
	prompt   string
}

type EvaluatorAgentConfiguration struct {
	Alert  string
	Result string
}

func (agent *EvaluatorAgent) Configure(agentConfiguration AgentConfiguration) {
	agent.aiClient = agentConfiguration.AiClient
	agent.context = agentConfiguration.Context
	agent.alert = agentConfiguration.Evaluator.Alert
	agent.result = agentConfiguration.Evaluator.Result
	agent.prompt = `
You are a Kubernetes expert. Your role is to assess whether an error detected in the cluster matches an error sent by the user. Your response is crucial in resolving the alert, so you must be VERY precise.

The alert : "%s"
The error : "%s"

IMPORTANT: You will respond with "true" only if the alert matches the error and "false" in all other cases.
	`
}

func (agent EvaluatorAgent) Process() (bool, error) {

	result := false
	response, err := agent.aiClient.GetCompletion(agent.context, fmt.Sprintf(agent.prompt, agent.alert, agent.result))
	if err != nil {
		return result, err
	}

	if strings.ToLower(response) == "true" {
		result = true
	}

	return result, nil
}
