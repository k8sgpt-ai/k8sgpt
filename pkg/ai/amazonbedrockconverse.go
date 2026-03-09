package ai

import (
	"context"
	"os"
	"fmt"
	"errors"
	"strings"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

const amazonBedrockConverseClientName = "amazonbedrockconverse"

// AmazonBedrockConverseClient represents the client for interacting with the Amazon Bedrock service.
type AmazonBedrockConverseClient struct {
	nopCloser

	client *bedrockruntime.Client
	model string
	// config types.InferenceConfiguration
	temperature float32
	topP float32
	maxTokens int
	stopSequences []string
}

func GetRegion(region string) string {
	if os.Getenv("AWS_DEFAULT_REGION") != "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}
	// Return the supplied provider region if not overridden by environment variable
	return region
}

// Get model from string
func (a *AmazonBedrockConverseClient) getModelFromString(model string) (string, error) {
	if model == "" {
		return "", errors.New("model name cannot be empty")
	}
	// Trim spaces from the model name
	model = strings.TrimSpace(model)

	return model, nil
}

func ProcessError(err error, modelId string) {
	errMsg := err.Error()
	if strings.Contains(errMsg, "no such host") {
		fmt.Errorf(`The Bedrock service is not available in the selected region.
                    Please double-check the service availability for your region at
                    https://aws.amazon.com/about-aws/global-infrastructure/regional-product-services/.\n`)
	} else if strings.Contains(errMsg, "Could not resolve the foundation model") {
		fmt.Errorf(`Could not resolve the foundation model from model identifier: \"%w\".
                    Please verify that the requested model exists and is accessible
                    within the specified region.\n
                    `, modelId)
	} else {
		fmt.Errorf("Couldn't invoke model: \"%w\". Here's why: %w\n", modelId, err)
	}
}

// Configure configures the AmazonBedrockConverseClient with the provided configuration.
func (a *AmazonBedrockConverseClient) Configure(config IAIConfig) error {
	// Get the model input
	modelInput := config.GetModel()

	// Determine the appropriate region to use
	var region = GetRegion(config.GetProviderRegion())

	// Only create AWS clients if they haven't been injected (for testing)
	if a.client == nil {
		// Create a new AWS config with the determined region
		cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion(region),
		)
		if err != nil {
			if strings.Contains(err.Error(), "InvalidAccessKeyId") || strings.Contains(err.Error(), "SignatureDoesNotMatch") || strings.Contains(err.Error(), "NoCredentialProviders") {
				return fmt.Errorf("AWS credentials are invalid or missing. Please check your AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables or AWS config. Details: %v", err)
			}
			return fmt.Errorf("failed to load AWS config for region %s: %w", region, err)
		}

		// Create clients with the config
		a.client = bedrockruntime.NewFromConfig(cfg)
	}

	foundModel, err := a.getModelFromString(modelInput)
		if err != nil {
			// Instead of failing, use a generic config for completion/response
			// But still warn user
			return fmt.Errorf("failed to find model configuration for %w: %w", modelInput, err)
		}
	// Use the found model config for completion/response, but set ModelName to the input ARN
	a.model = foundModel

	// Set common configuration parameters
	a.temperature = config.GetTemperature()
	a.topP = config.GetTopP()
	a.maxTokens = config.GetMaxTokens()
	a.stopSequences = config.GetStopSequences()

	return nil
}

// GetCompletion sends a request to the model for generating completion based on the provided prompt.
func (a *AmazonBedrockConverseClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	var content = types.ContentBlockMemberText{
		Value: prompt,
	}
	var message = types.Message{
		Content: []types.ContentBlock{&content},
		Role:    "user",
	}
	var converseInput = bedrockruntime.ConverseInput{
		ModelId:  aws.String(a.model),
		Messages: []types.Message{message},
		InferenceConfig: &types.InferenceConfiguration{
			Temperature: aws.Float32(a.temperature),
			TopP:       aws.Float32(a.topP),
			MaxTokens:  aws.Int32(int32(a.maxTokens)),
			StopSequences: a.stopSequences,
		},
	}
	response, err := a.client.Converse(ctx, &converseInput)
	if err != nil {
		ProcessError(err, a.model)
	}

	responseText, _ := response.Output.(*types.ConverseOutputMemberMessage)
	responseContentBlock := responseText.Value.Content[0]
	text, _ := responseContentBlock.(*types.ContentBlockMemberText)
	return text.Value, nil
}

// GetName returns the name of the AmazonBedrockConverseClient.
func (a *AmazonBedrockConverseClient) GetName() string {
	return amazonBedrockConverseClientName
}
