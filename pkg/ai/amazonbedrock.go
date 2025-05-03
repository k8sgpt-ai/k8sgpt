package ai

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai/bedrock_support"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

const amazonbedrockAIClientName = "amazonbedrock"

// AmazonBedRockClient represents the client for interacting with the Amazon Bedrock service.
type AmazonBedRockClient struct {
	nopCloser

	client      *bedrockruntime.Client
	mgmtClient  *bedrock.Client
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

var defaultModels = []bedrock_support.BedrockModel{
	{
		Name:       "us.anthropic.claude-3-7-sonnet-20250219-v1:0",
		Completion: &bedrock_support.CohereMessagesCompletion{},
		Response:   &bedrock_support.CohereMessagesResponse{},
		Config: bedrock_support.BedrockModelConfig{
			// sensible defaults
			MaxTokens:   100,
			Temperature: 0.5,
			TopP:        0.9,
			ModelName:   "us.anthropic.claude-3-7-sonnet-20250219-v1:0",
		},
	},
	{
		Name:       "anthropic.claude-3-5-sonnet-20240620-v1:0",
		Completion: &bedrock_support.CohereCompletion{},
		Response:   &bedrock_support.CohereResponse{},
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
			ModelName:   "eu.amazon.nova-pro-v1:0",
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

// NewAmazonBedRockClient creates a new AmazonBedRockClient with the given models
func NewAmazonBedRockClient(models []bedrock_support.BedrockModel) *AmazonBedRockClient {
	if models == nil {
		models = defaultModels // Use default models if none provided
	}
	return &AmazonBedRockClient{
		models: models,
	}
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

func validateModelArn(model string) bool {
	var re = regexp.MustCompile(`(?m)^arn:(?P<Partition>[^:\n]*):bedrock:(?P<Region>[^:\n]*):(?P<AccountID>[^:\n]*):(?P<Ignore>(?P<ResourceType>[^:\/\n]*)[:\/])?(?P<Resource>.*)$`)
	return re.MatchString(model)
}

func validateInferenceProfileArn(inferenceProfile string) bool {
	// Support both inference-profile and application-inference-profile formats
	var re = regexp.MustCompile(`(?m)^arn:(?P<Partition>[^:\n]*):bedrock:(?P<Region>[^:\n]*):(?P<AccountID>[^:\n]*):(?:inference-profile|application-inference-profile)\/(?P<ProfileName>.+)$`)
	return re.MatchString(inferenceProfile)
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

// Configure configures the AmazonBedRockClient with the provided configuration.
func (a *AmazonBedRockClient) Configure(config IAIConfig) error {
	// Initialize models if not already initialized
	if a.models == nil {
		a.models = defaultModels
	}

	// Get the model input
	modelInput := config.GetModel()
	
	// Determine the appropriate region to use
	var region string
	
	// Check if the model input is actually an inference profile ARN
	if validateInferenceProfileArn(modelInput) {
		// Extract the region from the inference profile ARN
		arnParts := strings.Split(modelInput, ":")
		if len(arnParts) >= 4 {
			region = arnParts[3]
		} else {
			return fmt.Errorf("could not extract region from inference profile ARN: %s", modelInput)
		}
	} else {
		// Use the provided region or default
		region = GetRegionOrDefault(config.GetProviderRegion())
	}
	
	// Create a new AWS config with the determined region
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), 
		awsconfig.WithRegion(region),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config for region %s: %w", region, err)
	}

	// Create clients with the config
	a.client = bedrockruntime.NewFromConfig(cfg)
	a.mgmtClient = bedrock.NewFromConfig(cfg)
	
	// Handle model selection based on input type
	if validateInferenceProfileArn(modelInput) {
		// Get the inference profile details
		profile, err := a.getInferenceProfile(context.Background(), modelInput)
		if err != nil {
			// Instead of using a fallback model, throw an error
			return fmt.Errorf("failed to get inference profile: %v", err)
		} else {
			// Extract the model ID from the inference profile
			modelID, err := a.extractModelFromInferenceProfile(profile)
			if err != nil {
				return fmt.Errorf("failed to extract model ID from inference profile: %v", err)
			}
			
			// Find the model configuration for the extracted model ID
			foundModel, err := a.getModelFromString(modelID)
			if err != nil {
				// Instead of using a fallback model, throw an error
				return fmt.Errorf("failed to find model configuration for %s: %v", modelID, err)
			}
			a.model = foundModel
			}
			
			// Use the inference profile ARN as the model ID for API calls
			a.model.Config.ModelName = modelInput
		}
	} else {
		// Regular model ID provided
		foundModel, err := a.getModelFromString(modelInput)
		if err != nil {
			return err
		}
		a.model = foundModel
		a.model.Config.ModelName = foundModel.Config.ModelName
	}
	
	// Set common configuration parameters
	a.temperature = config.GetTemperature()
	a.topP = config.GetTopP()
	a.maxTokens = config.GetMaxTokens()

	return nil
}

// getInferenceProfile retrieves the inference profile details from Amazon Bedrock
func (a *AmazonBedRockClient) getInferenceProfile(ctx context.Context, inferenceProfileARN string) (*bedrock.GetInferenceProfileOutput, error) {
	// Extract the profile ID from the ARN
	// ARN format: arn:aws:bedrock:region:account-id:inference-profile/profile-id
	// or arn:aws:bedrock:region:account-id:application-inference-profile/profile-id
	parts := strings.Split(inferenceProfileARN, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid inference profile ARN format: %s", inferenceProfileARN)
	}
	
	profileID := parts[1]
	
	// Create the input for the GetInferenceProfile API call
	input := &bedrock.GetInferenceProfileInput{
		InferenceProfileIdentifier: aws.String(profileID),
	}
	
	// Call the GetInferenceProfile API
	output, err := a.mgmtClient.GetInferenceProfile(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get inference profile: %w", err)
	}
	
	return output, nil
}

// extractModelFromInferenceProfile extracts the model ID from the inference profile
func (a *AmazonBedRockClient) extractModelFromInferenceProfile(profile *bedrock.GetInferenceProfileOutput) (string, error) {
	if profile == nil || len(profile.Models) == 0 {
		return "", fmt.Errorf("inference profile does not contain any models")
	}
	
	// Check if the first model exists and has a non-nil ModelArn
	if profile.Models[0] == nil || profile.Models[0].ModelArn == nil {
		return "", fmt.Errorf("model information is missing in inference profile")
	}
	
	// Get the first model ARN from the profile
	modelARN := aws.ToString(profile.Models[0].ModelArn)
	if modelARN == "" {
		return "", fmt.Errorf("model ARN is empty in inference profile")
	}
	
	// Extract the model ID from the ARN
	// ARN format: arn:aws:bedrock:region::foundation-model/model-id
	parts := strings.Split(modelARN, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid model ARN format: %s", modelARN)
	}
	
	modelID := parts[1]
	return modelID, nil
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
	resp, err := a.client.InvokeModel(ctx, params)
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
