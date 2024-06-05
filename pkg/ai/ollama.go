package ai

const ollamaClientName = "ollama"

type OllamaClient struct {
	OpenAIClient
}

func (a *OllamaClient) GetName() string {
	return ollamaClientName
}
