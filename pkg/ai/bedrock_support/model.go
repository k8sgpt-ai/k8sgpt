package bedrock_support

type BedrockModelConfig struct {
	MaxTokens           int
	Temperature         float32
	TopP                float32
	ModelName           string
}
type BedrockModel struct {
	Name       string
	Completion ICompletion
	Response   IResponse
	Config     BedrockModelConfig
}
