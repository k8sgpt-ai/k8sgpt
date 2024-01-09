package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
)

type RouterAgent struct {
	aiClient  ai.IAI
	context   context.Context
	alert     string
	analyzers []string
	prompt    string
}

type RouterAgentConfiguration struct {
	Alert     string
	Analyzers []string
}

type RouterResponse struct {
	SelectedAnalyzers []string `json:"selectedAnalyzers"`
}

func (agent *RouterAgent) Configure(agentConfiguration AgentConfiguration) {
	agent.aiClient = agentConfiguration.AiClient
	agent.alert = agentConfiguration.Router.Alert
	agent.analyzers = agentConfiguration.Router.Analyzers
	agent.context = agentConfiguration.Context
	agent.prompt = `You are part of a multi-agent autonomous AI system. You are Kubernetes expert and are responsible for assisting SREs in understanding and resolving issues. The system consists of analyzers. Each analyzer can be used to analyze a specific type of resource or function in a Kubernetes cluster. It contains an error identification logic. Your role is to decide which analyzers I should activate to investigate the following alert:
	Error: %s
	
	The list of all analyzers here :

%s

	Feel free to return multiple analyzers in order to obtain a comprehensive understanding of the error! The result cannot be empty.

	WARNING: Your response must ONLY contain a JSON table containing the names of the selected analyzers as strings. No other format is allowed.

	Full Response example:

	{
		"selectedAnalyzers": ["Service", "Pod"]
	}
	`
}

func (agent *RouterAgent) Process() ([]string, error) {
	var analyzerString strings.Builder
	for i, element := range agent.analyzers {
		analyzerString.WriteString(fmt.Sprintf("%d. %s\n", i+1, element))
	}

	//fmt.Println(fmt.Sprintf(agent.prompt, agent.alert, analyzerString.String()))
	response, err := agent.aiClient.GetCompletion(agent.context, fmt.Sprintf(agent.prompt, agent.alert, analyzerString.String()))
	if err != nil {
		return []string{}, err
	}

	data := RouterResponse{}

	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return []string{}, err
	}

	return data.SelectedAnalyzers, nil
}
