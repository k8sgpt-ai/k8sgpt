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

type AmazonBedrockConverseClient struct {
	nopCloser

	client *bedrockruntime.Client
	model string
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

func (a *AmazonBedrockConverseClient) getModelFromString(model string) (string, error) {
	if model == "" {
		return "", errors.New("model name cannot be empty")
	}
	model = strings.TrimSpace(model)

	return model, nil
}

func ProcessError(err error, modelId string) error {
	errMsg := err.Error()
	if strings.Contains(errMsg, "no such host") {
		return fmt.Errorf(`The Bedrock service is not available in the selected region.
                    Please double-check the service availability for your region at
                    https://aws.amazon.com/about-aws/global-infrastructure/regional-product-services/.\n`)
	} else if strings.Contains(errMsg, "Could not resolve the foundation model") {
		return fmt.Errorf(`Could not resolve the foundation model from model identifier: \"%s\".
                    Please verify that the requested model exists and is accessible
                    within the specified region.\n
                    `, modelId)
	} else {
		return fmt.Errorf("Couldn't invoke model: \"%s\". Here's why: %s\n", modelId, err)
	}
}

func (a *AmazonBedrockConverseClient) Configure(config IAIConfig) error {
	modelInput := config.GetModel()

	var region = GetRegion(config.GetProviderRegion())

	// Only create AWS clients if they haven't been injected (for testing)
	if a.client == nil {
		cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion(region),
		)
		if err != nil {
			if strings.Contains(err.Error(), "InvalidAccessKeyId") || strings.Contains(err.Error(), "SignatureDoesNotMatch") || strings.Contains(err.Error(), "NoCredentialProviders") {
				return fmt.Errorf("AWS credentials are invalid or missing. Please check your AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables or AWS config. Details: %v", err)
			}
			return fmt.Errorf("failed to load AWS config for region %s: %w", region, err)
		}

		a.client = bedrockruntime.NewFromConfig(cfg)
	}

	foundModel, err := a.getModelFromString(modelInput)
		if err != nil {
			return fmt.Errorf("failed to find model configuration for %s: %w", modelInput, err)
		}
	a.model = foundModel

	// Set common configuration parameters
	a.temperature = config.GetTemperature()
	a.topP = config.GetTopP()
	a.maxTokens = config.GetMaxTokens()
	a.stopSequences = config.GetStopSequences()

	return nil
}

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
		return "",ProcessError(err, a.model)
	}

	responseText, _ := response.Output.(*types.ConverseOutputMemberMessage)
	responseContentBlock := responseText.Value.Content[0]
	text, _ := responseContentBlock.(*types.ContentBlockMemberText)
	return text.Value, nil
}

func (a *AmazonBedrockConverseClient) GetName() string {
	return amazonBedrockConverseClientName
}
