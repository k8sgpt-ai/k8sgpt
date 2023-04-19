package ai

type LLaMAAIClient struct {
	OpenAIClient
}

func (a *LLaMAAIClient) GetName() string {
	return "llama"
}
