package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"

	"github.com/k8sgpt-ai/k8sgpt/pkg/cache"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/bedrockruntime"
)

// AmazonBedRockClient represents the client for interacting with the Amazon Bedrock service.
type AmazonBedRockClient struct {
	client      *bedrockruntime.BedrockRuntime
	language    string
	model       string
	temperature float32
}

// InvokeModelResponseBody represents the response body structure from the model invocation.
type InvokeModelResponseBody struct {
	Completion  string `json:"completion"`
	Stop_reason string `json:"stop_reason"`
}

const BEDROCK_REGION = "us-east-1" // default use us-east-1 region

const (
	ModelAnthropicClaudeV2        = "anthropic.claude-v2"
	ModelAnthropicClaudeV1        = "anthropic.claude-v1"
	ModelAnthropicClaudeInstantV1 = "anthropic.claude-instant-v1"
)

var BEDROCK_MODELS = []string{
	ModelAnthropicClaudeV2,
	ModelAnthropicClaudeV1,
	ModelAnthropicClaudeInstantV1,
}

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

// Configure configures the AmazonBedRockClient with the provided configuration and language.
func (a *AmazonBedRockClient) Configure(config IAIConfig, language string) error {

	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(BEDROCK_REGION),
	})

	if err != nil {
		return err
	}

	// Create a new BedrockRuntime client
	a.client = bedrockruntime.New(sess)
	a.language = language
	a.model = GetModelOrDefault(config.GetModel())
	a.temperature = config.GetTemperature()

	return nil
}

// GetCompletion sends a request to the model for generating completion based on the provided prompt.
func (a *AmazonBedRockClient) GetCompletion(ctx context.Context, prompt string, promptTmpl string) (string, error) {

	// Prepare the input data for the model invocation
	request := map[string]interface{}{
		"prompt":               fmt.Sprintf("\n\nHuman: %s  \n\nAssistant:", prompt),
		"max_tokens_to_sample": 1024,
		"temperature":          a.temperature,
		"top_p":                0.9,
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
	// Parse the response body
	output := &InvokeModelResponseBody{}
	err = json.Unmarshal(resp.Body, output)
	if err != nil {
		return "", err
	}
	return output.Completion, nil
}

// Parse generates a completion for the provided prompt using the Amazon Bedrock model.
func (a *AmazonBedRockClient) Parse(ctx context.Context, prompt []string, cache cache.ICache, promptTmpl string) (string, error) {
	inputKey := strings.Join(prompt, " ")
	// Check for cached data
	cacheKey := util.GetCacheKey(a.GetName(), a.language, inputKey)

	if !cache.IsCacheDisabled() && cache.Exists(cacheKey) {
		response, err := cache.Load(cacheKey)
		if err != nil {
			return "", err
		}

		if response != "" {
			output, err := base64.StdEncoding.DecodeString(response)
			if err != nil {
				color.Red("error decoding cached data: %v", err)
				return "", nil
			}
			return string(output), nil
		}
	}

	response, err := a.GetCompletion(ctx, inputKey, promptTmpl)

	if err != nil {
		return "", err
	}

	err = cache.Store(cacheKey, base64.StdEncoding.EncodeToString([]byte(response)))

	if err != nil {
		color.Red("error storing value to cache: %v", err)
		return "", nil
	}

	return response, nil
}

// GetName returns the name of the AmazonBedRockClient.
func (a *AmazonBedRockClient) GetName() string {
	return "amazonbedrock"
}
