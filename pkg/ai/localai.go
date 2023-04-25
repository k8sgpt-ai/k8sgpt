package ai

type LocalAIClient struct {
	OpenAIClient
}

func (a *LocalAIClient) GetName() string {
	return "localai"
}
