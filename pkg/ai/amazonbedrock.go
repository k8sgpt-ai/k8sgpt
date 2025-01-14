package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/bedrockruntime"
)

const amazonbedrockAIClientName = "amazonbedrock"

// AmazonBedRockClient represents the client for interacting with the Amazon Bedrock service.
type AmazonBedRockClient struct {
	nopCloser

	client      *bedrockruntime.BedrockRuntime
	model       string
	temperature float32
	topP        float32
	maxTokens   int
}

// Amazon BedRock support region list US East (N. Virginia),US West (Oregon),Asia Pacific (Singapore),Asia Pacific (Tokyo),Europe (Frankfurt)
// https://docs.aws.amazon.com/bedrock/latest/userguide/what-is-bedrock.html#bedrock-regions
const BEDROCK_DEFAULT_REGION = "us-east-1" // default use us-east-1 region

const (
	US_East_1      = "us-east-1"
	US_West_2      = "us-west-2"
	AP_Southeast_1 = "ap-southeast-1"
	AP_Northeast_1 = "ap-northeast-1"
	EU_Central_1   = "eu-central-1"
)

var BEDROCKER_SUPPORTED_REGION = []string{
	US_East_1,
	US_West_2,
	AP_Southeast_1,
	AP_Northeast_1,
	EU_Central_1,
}

const (
	ModelAnthropicClaudeSonnetV3_5    = "anthropic.claude-3-5-sonnet-20240620-v1:0"
	ModelAnthropicClaudeSonnetV3_5_V2 = "us.anthropic.claude-3-5-sonnet-20241022-v2:0"
	ModelAnthropicClaudeV2            = "anthropic.claude-v2"
	ModelAnthropicClaudeV1            = "anthropic.claude-v1"
	ModelAnthropicClaudeInstantV1     = "anthropic.claude-instant-v1"
	ModelA21J2UltraV1                 = "ai21.j2-ultra-v1"
	ModelA21J2JumboInstruct           = "ai21.j2-jumbo-instruct"
	ModelAmazonTitanExpressV1         = "amazon.titan-text-express-v1"
)

var BEDROCK_MODELS = []string{
	ModelAnthropicClaudeV2,
	ModelAnthropicClaudeV1,
	ModelAnthropicClaudeInstantV1,
	ModelA21J2UltraV1,
	ModelA21J2JumboInstruct,
	ModelAmazonTitanExpressV1,
}

//const TOPP = 0.9 moved to config

// GetModelOrDefault check config model
func GetModelOrDefault(model string) string {

	// Check if the provided model is in the list
	for _, m := range BEDROCK_MODELS {
		if m == model {
			return model // Return the provided model
		}
	}

	// Return the default model if the provided model is not in the list
	return BEDROCK_MODELS[0]
}

// GetModelOrDefault check config region
func GetRegionOrDefault(region string) string {

	if os.Getenv("AWS_DEFAULT_REGION") != "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}
	// Check if the provided model is in the list
	for _, m := range BEDROCKER_SUPPORTED_REGION {
		if m == region {
			return region // Return the provided model
		}
	}

	// Return the default model if the provided model is not in the list
	return BEDROCK_DEFAULT_REGION
}

// Configure configures the AmazonBedRockClient with the provided configuration.
func (a *AmazonBedRockClient) Configure(config IAIConfig) error {

	// Create a new AWS session
	providerRegion := GetRegionOrDefault(config.GetProviderRegion())

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(providerRegion),
	})

	if err != nil {
		return err
	}

	// Create a new BedrockRuntime client
	a.client = bedrockruntime.New(sess)
	a.model = GetModelOrDefault(config.GetModel())
	a.temperature = config.GetTemperature()
	a.topP = config.GetTopP()
	a.maxTokens = config.GetMaxTokens()

	return nil
}

// GetCompletion sends a request to the model for generating completion based on the provided prompt.
func (a *AmazonBedRockClient) GetCompletion(ctx context.Context, prompt string) (string, error) {

	// Prepare the input data for the model invocation based on the model & the Response Body per model as well.
	var request map[string]interface{}
	switch a.model {
	case ModelAnthropicClaudeSonnetV3_5, ModelAnthropicClaudeSonnetV3_5_V2, ModelAnthropicClaudeV2, ModelAnthropicClaudeV1, ModelAnthropicClaudeInstantV1:
		request = map[string]interface{}{
			"prompt":               fmt.Sprintf("\n\nHuman: %s  \n\nAssistant:", prompt),
			"max_tokens_to_sample": a.maxTokens,
			"temperature":          a.temperature,
			"top_p":                a.topP,
		}
	case ModelA21J2UltraV1, ModelA21J2JumboInstruct:
		request = map[string]interface{}{
			"prompt":      prompt,
			"maxTokens":   a.maxTokens,
			"temperature": a.temperature,
			"topP":        a.topP,
		}
	case ModelAmazonTitanExpressV1:
		request = map[string]interface{}{
			"inputText": fmt.Sprintf("\n\nUser: %s", prompt),
			"textGenerationConfig": map[string]interface{}{
				"maxTokenCount": a.maxTokens,
				"temperature":   a.temperature,
				"topP":          a.topP,
			},
		}
	default:
		return "", fmt.Errorf("model %s not supported", a.model)
	}

	body, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	// Build the parameters for the model invocation
	params := &bedrockruntime.InvokeModelInput{
		Body:        body,
		ModelId:     aws.String(a.model),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	}
	// Invoke the model
	resp, err := a.client.InvokeModelWithContext(ctx, params)

	if err != nil {
		return "", err
	}

	// Response type changes as per model
	switch a.model {
	case ModelAnthropicClaudeSonnetV3_5, ModelAnthropicClaudeSonnetV3_5_V2, ModelAnthropicClaudeV2, ModelAnthropicClaudeV1, ModelAnthropicClaudeInstantV1:
		type InvokeModelResponseBody struct {
			Completion  string `json:"completion"`
			Stop_reason string `json:"stop_reason"`
		}
		output := &InvokeModelResponseBody{}
		err = json.Unmarshal(resp.Body, output)
		if err != nil {
			return "", err
		}
		return output.Completion, nil
	case ModelA21J2UltraV1, ModelA21J2JumboInstruct:
		type Data struct {
			Text string `json:"text"`
		}
		type Completion struct {
			Data Data `json:"data"`
		}
		type InvokeModelResponseBody struct {
			Completions []Completion `json:"completions"`
		}
		output := &InvokeModelResponseBody{}
		err = json.Unmarshal(resp.Body, output)
		if err != nil {
			return "", err
		}
		return output.Completions[0].Data.Text, nil
	case ModelAmazonTitanExpressV1:
		type Result struct {
			TokenCount       int    `json:"tokenCount"`
			OutputText       string `json:"outputText"`
			CompletionReason string `json:"completionReason"`
		}
		type InvokeModelResponseBody struct {
			InputTextTokenCount int      `json:"inputTextTokenCount"`
			Results             []Result `json:"results"`
		}
		output := &InvokeModelResponseBody{}
		err = json.Unmarshal(resp.Body, output)
		if err != nil {
			return "", err
		}
		return output.Results[0].OutputText, nil
	default:
		return "", fmt.Errorf("model %s not supported", a.model)
	}
}

// GetName returns the name of the AmazonBedRockClient.
func (a *AmazonBedRockClient) GetName() string {
	return amazonbedrockAIClientName
}
