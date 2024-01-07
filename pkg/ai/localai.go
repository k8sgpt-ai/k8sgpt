package ai

const localAIClientName = "localai"

type LocalAIClient struct {
	OpenAIClient
}

func (a *LocalAIClient) GetName() string {
	return localAIClientName
}
