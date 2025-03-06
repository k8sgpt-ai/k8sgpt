package ai

import (
	"context"
	"errors"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai/bedrock_support"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/bedrockruntime"
)

const amazonbedrockAIClientName = "amazonbedrock"

// AmazonBedRockClient represents the client for interacting with the AmazonCompletion Bedrock service.
type AmazonBedRockClient struct {
	nopCloser

	client      *bedrockruntime.BedrockRuntime
	model       *bedrock_support.BedrockModel
	temperature float32
	topP        float32
	maxTokens   int
}

// AmazonCompletion BedRock support region list US East (N. Virginia),US West (Oregon),Asia Pacific (Singapore),Asia Pacific (Tokyo),Europe (Frankfurt)
// https://docs.aws.amazon.com/bedrock/latest/userguide/what-is-bedrock.html#bedrock-regions
const BEDROCK_DEFAULT_REGION = "us-east-1" // default use us-east-1 region

const (
	US_East_1      = "us-east-1"
	US_West_2      = "us-west-2"
	AP_Southeast_1 = "ap-southeast-1"
	AP_Northeast_1 = "ap-northeast-1"
	EU_Central_1   = "eu-central-1",
	AP_South_1     = "ap-south-1"
)

var BEDROCKER_SUPPORTED_REGION = []string{
	US_East_1,
	US_West_2,
	AP_Southeast_1,
	AP_Northeast_1,
	EU_Central_1,
	AP_South_1,
}

var (
	models = []bedrock_support.BedrockModel{
		{
			Name:       "anthropic.claude-3-5-sonnet-20240620-v1:0",
			Completion: &bedrock_support.CohereCompletion{},
			Response:   &bedrock_support.CohereResponse{},
			Config: bedrock_support.BedrockModelConfig{
				// sensible defaults
				MaxTokens:   100,
				Temperature: 0.5,
				TopP:        0.9,
			},
		},
		{
			Name:       "us.anthropic.claude-3-5-sonnet-20241022-v2:0",
			Completion: &bedrock_support.CohereCompletion{},
			Response:   &bedrock_support.CohereResponse{},
			Config: bedrock_support.BedrockModelConfig{
				// sensible defaults
				MaxTokens:   100,
				Temperature: 0.5,
				TopP:        0.9,
			},
		},
		{
			Name:       "us.anthropic.claude-3-5-sonnet-20241022-v2:0",
			Completion: &bedrock_support.CohereCompletion{},
			Response:   &bedrock_support.CohereResponse{},
			Config: bedrock_support.BedrockModelConfig{
				// sensible defaults
				MaxTokens:   100,
				Temperature: 0.5,
				TopP:        0.9,
			},
		},
		{
			Name:       "anthropic.claude-v2",
			Completion: &bedrock_support.CohereCompletion{},
			Response:   &bedrock_support.CohereResponse{},
			Config: bedrock_support.BedrockModelConfig{
				// sensible defaults
				MaxTokens:   100,
				Temperature: 0.5,
				TopP:        0.9,
			},
		},
		{
			Name:       "anthropic.claude-v1",
			Completion: &bedrock_support.CohereCompletion{},
			Response:   &bedrock_support.CohereResponse{},
			Config: bedrock_support.BedrockModelConfig{
				// sensible defaults
				MaxTokens:   100,
				Temperature: 0.5,
				TopP:        0.9,
			},
		},
		{
			Name:       "anthropic.claude-instant-v1",
			Completion: &bedrock_support.CohereCompletion{},
			Response:   &bedrock_support.CohereResponse{},
			Config: bedrock_support.BedrockModelConfig{
				// sensible defaults
				MaxTokens:   100,
				Temperature: 0.5,
				TopP:        0.9,
			},
		},
		{
			Name:       "ai21.j2-ultra-v1",
			Completion: &bedrock_support.AI21{},
			Response:   &bedrock_support.AI21Response{},
			Config: bedrock_support.BedrockModelConfig{
				// sensible defaults
				MaxTokens:   100,
				Temperature: 0.5,
				TopP:        0.9,
			},
		},
		{
			Name:       "ai21.j2-jumbo-instruct",
			Completion: &bedrock_support.AI21{},
			Response:   &bedrock_support.AI21Response{},
			Config: bedrock_support.BedrockModelConfig{
				// sensible defaults
				MaxTokens:   100,
				Temperature: 0.5,
				TopP:        0.9,
			},
		},
		{
			Name:       "amazon.titan-text-express-v1",
			Completion: &bedrock_support.AmazonCompletion{},
			Response:   &bedrock_support.AmazonResponse{},
			Config: bedrock_support.BedrockModelConfig{
				// sensible defaults
				MaxTokens:   100,
				Temperature: 0.5,
				TopP:        0.9,
			},
		},
	}
)

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

// Get model from string
func (a *AmazonBedRockClient) getModelFromString(model string) (*bedrock_support.BedrockModel, error) {
	for _, m := range models {
		if model == m.Name {
			return &m, nil
		}
	}
	return nil, errors.New("model not found")
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

	foundModel, err := a.getModelFromString(config.GetModel())
	if err != nil {
		return err
	}
	// TODO: Override the completion config somehow

	// Create a new BedrockRuntime client
	a.client = bedrockruntime.New(sess)
	a.model = foundModel
	a.temperature = config.GetTemperature()
	a.topP = config.GetTopP()
	a.maxTokens = config.GetMaxTokens()

	return nil
}

// GetCompletion sends a request to the model for generating completion based on the provided prompt.
func (a *AmazonBedRockClient) GetCompletion(ctx context.Context, prompt string) (string, error) {

	// override config defaults
	a.model.Config.MaxTokens = a.maxTokens
	a.model.Config.Temperature = a.temperature
	a.model.Config.TopP = a.topP

	body, err := a.model.Completion.GetCompletion(ctx, prompt, a.model.Config)
	if err != nil {
		return "", err
	}
	// Build the parameters for the model invocation
	params := &bedrockruntime.InvokeModelInput{
		Body:        body,
		ModelId:     aws.String(a.model.Name),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	}
	// Invoke the model
	resp, err := a.client.InvokeModelWithContext(ctx, params)

	if err != nil {
		return "", err
	}

	// Parse the response
	return a.model.Response.ParseResponse(resp.Body)

}

// GetName returns the name of the AmazonBedRockClient.
func (a *AmazonBedRockClient) GetName() string {
	return amazonbedrockAIClientName
}
