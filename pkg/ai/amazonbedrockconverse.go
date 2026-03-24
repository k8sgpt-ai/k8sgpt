package ai

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"os"
	"strings"
)

const amazonBedrockConverseClientName = "amazonbedrockconverse"

type bedrockConverseAPI interface {
	Converse(ctx context.Context, input *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error)
}

type AmazonBedrockConverseClient struct {
	nopCloser

	client        bedrockConverseAPI
	model         string
	temperature   float32
	topP          float32
	maxTokens     int
	stopSequences []string
}

func getRegion(region string) string {
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

func processError(err error, modelId string) error {
	errMsg := err.Error()
	if strings.Contains(errMsg, "no such host") {
		return fmt.Errorf(`the bedrock service is not available in the selected region.
                    please double-check the service availability for your region at
                    https://aws.amazon.com/about-aws/global-infrastructure/regional-product-services/`)
	} else if strings.Contains(errMsg, "Could not resolve the foundation model") {
		return fmt.Errorf(`could not resolve the foundation model from model identifier: \"%s\".
                    please verify that the requested model exists and is accessible
                    within the specified region`, modelId)
	} else {
		return fmt.Errorf("could not invoke model: \"%s\". here is why: %s", modelId, err)
	}
}

func isClaudeModel(modelId string) bool {
	m := strings.ToLower(modelId)
	return strings.Contains(m, "claude")
}

func (a *AmazonBedrockConverseClient) Configure(config IAIConfig) error {
	modelInput := config.GetModel()

	var region = getRegion(config.GetProviderRegion())

	// Only create AWS clients if they haven't been injected (for testing)
	if a.client == nil {
		cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion(region),
		)
		if err != nil {
			if strings.Contains(err.Error(), "InvalidAccessKeyId") || strings.Contains(err.Error(), "SignatureDoesNotMatch") || strings.Contains(err.Error(), "NoCredentialProviders") {
				return fmt.Errorf("aws credentials are invalid or missing. Please check your environment variables or aws config. details: %v", err)
			}
			return fmt.Errorf("failed to load aws config for region %s: %w", region, err)
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

func extractTextFromConverseOutput(output types.ConverseOutput, modelId string) (string, error) {
	if output == nil {
		return "", fmt.Errorf("empty response from model: %s", modelId)
	}

	msg, ok := output.(*types.ConverseOutputMemberMessage)
	if !ok {
		return "", fmt.Errorf("unexpected response type from model: %s", modelId)
	}

	if len(msg.Value.Content) == 0 {
		return "", fmt.Errorf("no content returned from model: %s", modelId)
	}

	var builder strings.Builder

	for _, block := range msg.Value.Content {
		if textBlock, ok := block.(*types.ContentBlockMemberText); ok && textBlock != nil {
			builder.WriteString(textBlock.Value)
		}
	}

	if builder.Len() == 0 {
		return "", fmt.Errorf("no text content returned from model: %s", modelId)
	}

	return builder.String(), nil
}

func (a *AmazonBedrockConverseClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	var content = types.ContentBlockMemberText{
		Value: prompt,
	}
	var message = types.Message{
		Content: []types.ContentBlock{&content},
		Role:    "user",
	}

	var infConfig = &types.InferenceConfiguration{
		MaxTokens:     aws.Int32(int32(a.maxTokens)),
		StopSequences: a.stopSequences,
	}

	// Claude models only support temperature OR topP, while others support both temperature and topP. Prefer temperature for now
	if !isClaudeModel(a.model) {
		infConfig.TopP = aws.Float32(a.topP)
	}
	infConfig.Temperature = aws.Float32(a.temperature)
		
	var converseInput = bedrockruntime.ConverseInput{
		ModelId:  aws.String(a.model),
		Messages: []types.Message{message},
		InferenceConfig: infConfig,
	}
	response, err := a.client.Converse(ctx, &converseInput)
	if err != nil {
		return "", processError(err, a.model)
	}

	text, err := extractTextFromConverseOutput(response.Output, a.model)
	if err != nil {
		return "", err
	}

	return text, nil
}

func (a *AmazonBedrockConverseClient) GetName() string {
	return amazonBedrockConverseClientName
}
