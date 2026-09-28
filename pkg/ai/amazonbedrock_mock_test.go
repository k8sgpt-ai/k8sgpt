package ai

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai/bedrock_support"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for Bedrock Management Client
type MockBedrockClient struct {
	mock.Mock
}

func (m *MockBedrockClient) GetInferenceProfile(ctx context.Context, params *bedrock.GetInferenceProfileInput, optFns ...func(*bedrock.Options)) (*bedrock.GetInferenceProfileOutput, error) {
	args := m.Called(ctx, params)
	
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	
	return args.Get(0).(*bedrock.GetInferenceProfileOutput), args.Error(1)
}

// Mock for Bedrock Runtime Client
type MockBedrockRuntimeClient struct {
	mock.Mock
}

func (m *MockBedrockRuntimeClient) InvokeModel(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error) {
	args := m.Called(ctx, params)
	
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	
	return args.Get(0).(*bedrockruntime.InvokeModelOutput), args.Error(1)
}

// TestBedrockInferenceProfileARNWithMocks tests the inference profile ARN validation with mocks
func TestBedrockInferenceProfileARNWithMocks(t *testing.T) {
	// Create test models
	testModels := []bedrock_support.BedrockModel{
		{
			Name:       "anthropic.claude-3-5-sonnet-20240620-v1:0",
			Completion: &bedrock_support.CohereMessagesCompletion{},
			Response:   &bedrock_support.CohereMessagesResponse{},
			Config: bedrock_support.BedrockModelConfig{
				MaxTokens:   100,
				Temperature: 0.5,
				TopP:        0.9,
				ModelName:   "anthropic.claude-3-5-sonnet-20240620-v1:0",
			},
		},
	}
	
	// Create a client with test models
	client := &AmazonBedRockClient{models: testModels}
	
	// Create mock clients
	mockMgmtClient := new(MockBedrockClient)
	mockRuntimeClient := new(MockBedrockRuntimeClient)
	
	// Inject mock clients into the AmazonBedRockClient
	client.mgmtClient = mockMgmtClient
	client.client = mockRuntimeClient
	
	// Test with a valid inference profile ARN
	inferenceProfileARN := "arn:aws:bedrock:us-east-1:123456789012:inference-profile/my-profile"
	
	// Setup mock response for GetInferenceProfile
	mockMgmtClient.On("GetInferenceProfile", mock.Anything, &bedrock.GetInferenceProfileInput{
		InferenceProfileIdentifier: aws.String("my-profile"),
	}).Return(&bedrock.GetInferenceProfileOutput{
		Models: []types.InferenceProfileModel{
			{
				ModelArn: aws.String("arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-3-5-sonnet-20240620-v1:0"),
			},
		},
	}, nil)
	
	// Configure the client with the inference profile ARN
	config := AIProvider{
		Model:          inferenceProfileARN,
		ProviderRegion: "us-east-1",
	}
	
	// Test the Configure method with the inference profile ARN
	err := client.Configure(&config)
	
	// Verify that the configuration was successful
	assert.NoError(t, err)
	assert.Equal(t, inferenceProfileARN, client.model.Config.ModelName)
	
	// Verify that the mock was called
	mockMgmtClient.AssertExpectations(t)
}
