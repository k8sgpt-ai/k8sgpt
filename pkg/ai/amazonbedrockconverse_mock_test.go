package ai

import (
	"context"
	"errors"
	"testing"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/stretchr/testify/assert"
)

// ---- Mock Wrapper ----
type mockConverseClient struct {
	converseFunc func(ctx context.Context, input *bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error)
}

func (m *mockConverseClient) Converse(ctx context.Context, input *bedrockruntime.ConverseInput, _ ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
	return m.converseFunc(ctx, input)
}

// ---- Tests ----
func TestGetCompletion_Success(t *testing.T) {
	mock := &mockConverseClient{
		converseFunc: func(ctx context.Context, input *bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error) {
			return &bedrockruntime.ConverseOutput{
				Output: &types.ConverseOutputMemberMessage{
					Value: types.Message{
						Content: []types.ContentBlock{
							&types.ContentBlockMemberText{
								Value: "mock response",
							},
						},
					},
				},
			}, nil
		},
	}

	client := &AmazonBedrockConverseClient{
		model: "test-model",
	}

	client.client = (*bedrockruntime.Client)(nil)
	result, err := mock.Converse(context.Background(), &bedrockruntime.ConverseInput{})

	assert.NoError(t, err)

	output := result.Output.(*types.ConverseOutputMemberMessage)
	text := output.Value.Content[0].(*types.ContentBlockMemberText)

	assert.Equal(t, "mock response", text.Value)
}

func TestGetCompletion_Error(t *testing.T) {
	mock := &mockConverseClient{
		converseFunc: func(ctx context.Context, input *bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error) {
			return nil, errors.New("some error")
		},
	}

	_, err := mock.Converse(context.Background(), &bedrockruntime.ConverseInput{})
	assert.Error(t, err)
}

func TestGetName(t *testing.T) {
	client := &AmazonBedrockConverseClient{}
	assert.Equal(t, "amazonbedrockconverse", client.GetName())
}

type fakeConfig struct {
	model         string
	region        string
	temperature   float32
	topP          float32
	maxTokens     int
	stopSequences []string
}

func (f *fakeConfig) GetModel() string {
	return f.model
}

func (f *fakeConfig) GetProviderRegion() string {
	return f.region
}

func (f *fakeConfig) GetTemperature() float32 {
	return f.temperature
}

func (f *fakeConfig) GetTopP() float32 {
	return f.topP
}

func (f *fakeConfig) GetMaxTokens() int {
	return f.maxTokens
}

func (f *fakeConfig) GetStopSequences() []string {
	return f.stopSequences
}
