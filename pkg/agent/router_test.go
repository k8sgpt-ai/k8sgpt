package agent

import (
	"context"
	"errors"
	"os"
	"slices"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
)

const (
	maxIterations = 10
)

// TODO(matthisholleville): Being able to test the agent with other AI backends.
func setupRouterTestContext(alert string) (*RouterAgent, error) {
	var aiProvider ai.AIProvider
	aiProvider.Name = "openai"
	aiProvider.Password = os.Getenv("OPENAI_API_KEY")

	if aiProvider.Password == "" {
		return nil, errors.New("The test setup cannot be completed as no API key is provided.")
	}

	aiProvider.Model = "gpt-3.5-turbo"
	aiClient := ai.NewClient(aiProvider.Name)
	if err := aiClient.Configure(&aiProvider); err != nil {
		return nil, err
	}
	agentConfiguration := AgentConfiguration{
		AiClient: aiClient,
		Context:  context.Background(),
		Router: RouterAgentConfiguration{
			Alert:     alert,
			Analyzers: analyzer.GetAnalyzerNames(),
		},
	}
	routerAgent := RouterAgent{}
	routerAgent.Configure(agentConfiguration)
	return &routerAgent, nil
}

func TestServiceAlert(t *testing.T) {
	router, err := setupRouterTestContext("The Kubernetes service satelit-app-1 doesn't have targets")
	if err != nil {
		t.Logf(err.Error())
		t.Skip()
	}
	success := 0
	for i := 0; i < maxIterations; i++ {
		response, _ := router.Process()
		if slices.Contains(response, "Service") {
			success++
		}
	}
	successRate := success * 100 / maxIterations
	if successRate <= 80 {
		t.Errorf("The agent's accuracy must be above 80 percent. Currently, it's at %d percent.", successRate)
		return
	}

	t.Logf("TestPodAlert success rate %d percent.", successRate)
}

func TestPodAlert(t *testing.T) {
	router, err := setupRouterTestContext("Kubernetes pod crash looping: /crashloop-pod-d57db968d-zk2v9")
	if err != nil {
		t.Logf(err.Error())
		t.Skip()
	}
	success := 0
	for i := 0; i < maxIterations; i++ {
		response, _ := router.Process()
		if slices.Contains(response, "Pod") || slices.Contains(response, "Deployment") {
			success++
		}
	}
	successRate := success * 100 / maxIterations
	if successRate <= 80 {
		t.Errorf("The agent's accuracy must be above 80 percent. Currently, it's at %d percent.", successRate)
		return
	}
	t.Logf("TestPodAlert success rate %d percent with %d iterations.", successRate, maxIterations)
}
