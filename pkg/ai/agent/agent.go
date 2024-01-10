package agent

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analysis"
	"github.com/pterm/pterm"
)

type AGENT_STATE int

const (
	E_RUNNING AGENT_STATE = iota
	E_EXITED              = iota
)

type AIAgent struct {
	config        *analysis.Analysis
	contextWindow []byte
	State         chan AGENT_STATE
}

func NewAIAgent(config *analysis.Analysis, contextWindow []byte) *AIAgent {
	return &AIAgent{
		config:        config,
		contextWindow: contextWindow,
		State:         make(chan AGENT_STATE),
	}
}

func (a *AIAgent) StartInteraction() {

	a.State <- E_RUNNING
	pterm.Println("Interactive mode enabled [type exit to close.]")
	for {

		query := pterm.DefaultInteractiveTextInput.WithMultiLine(false)
		queryString, err := query.Show()
		if err != nil {
			fmt.Println(err)
		}
		if queryString == "" {
			continue
		}
		if strings.Contains(queryString, "exit") {
			a.State <- E_EXITED
			continue
		}
		pterm.Println()
		contextWindow := fmt.Sprintf("Given the context %s %s", string(a.contextWindow),
			queryString)

		response, err := a.config.AIClient.GetCompletion(a.config.Context,
			contextWindow)
		if err != nil {
			color.Red("Error: %v", err)
			a.State <- E_EXITED
			continue
		}
		pterm.Println(response)
	}
}
