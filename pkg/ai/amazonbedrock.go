package ai

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/service/bedrockruntime/bedrockruntimeiface"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai/bedrock_support"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/bedrockruntime"
)

const amazonbedrockAIClientName = "amazonbedrock"

// AmazonBedRockClient represents the client for interacting with the AmazonCompletion Bedrock service.
type AmazonBedRockClient struct {
	nopCloser

	client      bedrockruntimeiface.BedrockRuntimeAPI
	model       *bedrock_support.BedrockModel
	temperature float32
	topP        float32
	maxTokens   int
	models      []bedrock_support.BedrockModel
}

// AmazonCompletion BedRock support region list US East (N. Virginia),US West (Oregon),Asia Pacific (Singapore),Asia Pacific (Tokyo),Europe (Frankfurt)
// https://docs.aws.amazon.com/bedrock/latest/userguide/what-is-bedrock.html#bedrock-regions
const BEDROCK_DEFAULT_REGION = "us-east-1" // default use us-east-1 region

const (
	US_East_1      = "us-east-1"
	US_West_2      = "us-west-2"
	AP_Southeast_1 = "ap-southeast-1"
	AP_Northeast_1 = "ap-northeast-1"
	EU_Central_1   = "eu-central-1"
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
	defaultModels = []bedrock_support.BedrockModel{
		{
			Name:       "anthropic.claude-3-5-sonnet-20240620-v1:0",
			Completion: &bedrock_support.CohereMessagesCompletion{},
			Response:   &bedrock_support.CohereMessagesResponse{},
			Config: bedrock_support.BedrockModelConfig{
				// sensible defaults
				MaxTokens:   100,
				Temperature: 0.5,
				TopP:        0.9,
				ModelName:   "anthropic.claude-3-5-sonnet-20240620-v1:0",
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
				ModelName:   "us.anthropic.claude-3-5-sonnet-20241022-v2:0",
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
				ModelName:   "anthropic.claude-v2",
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
				ModelName:   "anthropic.claude-v1",
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
				ModelName:   "anthropic.claude-instant-v1",
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
				ModelName:   "ai21.j2-ultra-v1",
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
				ModelName:   "ai21.j2-jumbo-instruct",
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
				ModelName:   "amazon.titan-text-express-v1",
			},
		},
		{
			Name:       "amazon.nova-pro-v1:0",
			Completion: &bedrock_support.AmazonCompletion{},
			Response:   &bedrock_support.NovaResponse{},
			Config: bedrock_support.BedrockModelConfig{
				// sensible defaults
				// https://docs.aws.amazon.com/nova/latest/userguide/getting-started-api.html
				MaxTokens:   100, // max of 300k tokens
				Temperature: 0.5,
				TopP:        0.9,
				ModelName:   "amazon.nova-pro-v1:0",
			},
		},
		{
			Name:       "eu.amazon.nova-pro-v1:0",
			Completion: &bedrock_support.AmazonCompletion{},
			Response:   &bedrock_support.NovaResponse{},
			Config: bedrock_support.BedrockModelConfig{
				// sensible defaults
				// https://docs.aws.amazon.com/nova/latest/userguide/getting-started-api.html
				MaxTokens:   100, // max of 300k tokens
				Temperature: 0.5,
				TopP:        0.9,
				ModelName:   "eu.wamazon.nova-pro-v1:0",
			},
		},
		{
			Name:       "us.amazon.nova-pro-v1:0",
			Completion: &bedrock_support.AmazonCompletion{},
			Response:   &bedrock_support.NovaResponse{},
			Config: bedrock_support.BedrockModelConfig{
				// sensible defaults
				// https://docs.aws.amazon.com/nova/latest/userguide/getting-started-api.html
				MaxTokens:   100, // max of 300k tokens
				Temperature: 0.5,
				TopP:        0.9,
				ModelName:   "us.amazon.nova-pro-v1:0",
			},
		},
		{
			Name:       "amazon.nova-lite-v1:0",
			Completion: &bedrock_support.AmazonCompletion{},
			Response:   &bedrock_support.NovaResponse{},
			Config: bedrock_support.BedrockModelConfig{
				// sensible defaults
				MaxTokens:   100, // max of 300k tokens
				Temperature: 0.5,
				TopP:        0.9,
				ModelName:   "amazon.nova-lite-v1:0",
			},
		},
		{
			Name:       "eu.amazon.nova-lite-v1:0",
			Completion: &bedrock_support.AmazonCompletion{},
			Response:   &bedrock_support.NovaResponse{},
			Config: bedrock_support.BedrockModelConfig{
				// sensible defaults
				MaxTokens:   100, // max of 300k tokens
				Temperature: 0.5,
				TopP:        0.9,
				ModelName:   "eu.amazon.nova-lite-v1:0",
			},
		},
		{
			Name:       "us.amazon.nova-lite-v1:0",
			Completion: &bedrock_support.AmazonCompletion{},
			Response:   &bedrock_support.NovaResponse{},
			Config: bedrock_support.BedrockModelConfig{
				// sensible defaults
				MaxTokens:   100, // max of 300k tokens
				Temperature: 0.5,
				TopP:        0.9,
				ModelName:   "us.amazon.nova-lite-v1:0",
			},
		},
		{
			Name:       "anthropic.claude-3-haiku-20240307-v1:0",
			Completion: &bedrock_support.CohereCompletion{},
			Response:   &bedrock_support.CohereResponse{},
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
	if model == "" {
		return nil, errors.New("model name cannot be empty")
	}

	// Trim spaces from the model name
	model = strings.TrimSpace(model)
	modelLower := strings.ToLower(model)

	// Try to find an exact match first
	for i := range a.models {
		if strings.EqualFold(model, a.models[i].Name) || strings.EqualFold(model, a.models[i].Config.ModelName) {
			// Create a copy to avoid returning a pointer to a loop variable
			modelCopy := a.models[i]
			return &modelCopy, nil
		}
	}

	// If no exact match, try partial match
	for i := range a.models {
		modelNameLower := strings.ToLower(a.models[i].Name)
		modelConfigNameLower := strings.ToLower(a.models[i].Config.ModelName)

		// Check if the input string contains the model name or vice versa
		if strings.Contains(modelNameLower, modelLower) || strings.Contains(modelLower, modelNameLower) ||
			strings.Contains(modelConfigNameLower, modelLower) || strings.Contains(modelLower, modelConfigNameLower) {
			// Create a copy to avoid returning a pointer to a loop variable
			modelCopy := a.models[i]
			// for partial match, set the model name to the input string if it is a valid ARN
			if validateModelArn(modelLower) {
				modelCopy.Config.ModelName = modelLower
			}

			return &modelCopy, nil
		}
	}

	return nil, fmt.Errorf("model '%s' not found in supported models", model)
}

func validateModelArn(model string) bool {
	var re = regexp.MustCompile(`(?m)^arn:(?P<Partition>[^:\n]*):bedrock:(?P<Region>[^:\n]*):(?P<AccountID>[^:\n]*):(?P<Ignore>(?P<ResourceType>[^:\/\n]*)[:\/])?(?P<Resource>.*)$`)
	return re.MatchString(model)
}

// Configure configures the AmazonBedRockClient with the provided configuration.
func (a *AmazonBedRockClient) Configure(config IAIConfig) error {
	// Initialize models if not already initialized
	if a.models == nil {
		a.models = defaultModels
	}

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
	a.model.Config.ModelName = foundModel.Config.ModelName
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
		ModelId:     aws.String(a.model.Config.ModelName),
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
