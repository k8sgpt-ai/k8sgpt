package ai

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

// ---- Mock Wrapper ----
type mockConverseClient struct {
	converseFunc func(ctx context.Context, input *bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error)
	lastInput    *bedrockruntime.ConverseInput
}

func (m *mockConverseClient) Converse(ctx context.Context, input *bedrockruntime.ConverseInput, _ ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error) {
	m.lastInput = input
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
		client: mock,
		model:  "test-model",
	}

	result, err := client.GetCompletion(context.Background(), "hello")

	assert.NoError(t, err)
	assert.Equal(t, "mock response", result)
}

func TestGetCompletion_Error(t *testing.T) {
	mock := &mockConverseClient{
		converseFunc: func(ctx context.Context, input *bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error) {
			return nil, errors.New("some error")
		},
	}

	client := &AmazonBedrockConverseClient{
		client: mock,
		model:  "test-model",
	}

	_, err := client.GetCompletion(context.Background(), "hello")

	assert.Error(t, err)
}

func TestConfigure_WithInjectedClient(t *testing.T) {
	mock := &mockConverseClient{}

	cfg := &AIProvider{
		Model:          "test-model",
		ProviderRegion: "us-west-2",
		Temperature:    0.5,
		TopP:           0.9,
		MaxTokens:      100,
		StopSequences:  []string{"stop"},
	}

	client := &AmazonBedrockConverseClient{
		client: mock,
	}

	err := client.Configure(cfg)

	assert.NoError(t, err)
	assert.Equal(t, "test-model", client.model)
	assert.Equal(t, float32(0.5), client.temperature)
	assert.Equal(t, float32(0.9), client.topP)
	assert.Equal(t, 100, client.maxTokens)
	assert.Equal(t, []string{"stop"}, client.stopSequences)
}

func TestConfigure_InvalidModel(t *testing.T) {
	mock := &mockConverseClient{}

	cfg := &AIProvider{
		Model: "",
	}

	client := &AmazonBedrockConverseClient{
		client: mock,
	}

	err := client.Configure(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model name cannot be empty")
}

func TestGetRegion(t *testing.T) {
	t.Run("uses provided region when env not set", func(t *testing.T) {
		t.Setenv("AWS_DEFAULT_REGION", "")

		result := getRegion("us-west-2")
		assert.Equal(t, "us-west-2", result)
	})

	t.Run("env overrides provided region", func(t *testing.T) {
		t.Setenv("AWS_DEFAULT_REGION", "us-east-1")

		result := getRegion("us-west-2")
		assert.Equal(t, "us-east-1", result)
	})
}

func TestProcessError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		modelId  string
		contains string
	}{
		{
			name:     "no such host",
			err:      errors.New("dial tcp: no such host"),
			modelId:  "test-model",
			contains: "bedrock service is not available",
		},
		{
			name:     "model not found",
			err:      errors.New("Could not resolve the foundation model"),
			modelId:  "test-model",
			contains: "could not resolve the foundation model",
		},
		{
			name:     "generic error",
			err:      errors.New("something else"),
			modelId:  "test-model",
			contains: "could not invoke model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processError(tt.err, tt.modelId)
			assert.Contains(t, result.Error(), tt.contains)
		})
	}
}

func TestExtractTextFromConverseOutput(t *testing.T) {
	tests := []struct {
		name        string
		output      types.ConverseOutput
		expectError bool
		expected    string
	}{
		{
			name:        "nil output",
			output:      nil,
			expectError: true,
		},
		{
			name: "empty content",
			output: &types.ConverseOutputMemberMessage{
				Value: types.Message{
					Content: []types.ContentBlock{},
				},
			},
			expectError: true,
		},
		{
			name: "single text block",
			output: &types.ConverseOutputMemberMessage{
				Value: types.Message{
					Content: []types.ContentBlock{
						&types.ContentBlockMemberText{Value: "hello"},
					},
				},
			},
			expected: "hello",
		},
		{
			name: "multiple text blocks",
			output: &types.ConverseOutputMemberMessage{
				Value: types.Message{
					Content: []types.ContentBlock{
						&types.ContentBlockMemberText{Value: "hello "},
						&types.ContentBlockMemberText{Value: "world"},
					},
				},
			},
			expected: "hello world",
		},
		{
			name: "mixed content blocks",
			output: &types.ConverseOutputMemberMessage{
				Value: types.Message{
					Content: []types.ContentBlock{
						&types.ContentBlockMemberText{Value: "hello"},
						// simulate non-text block
						&types.ContentBlockMemberImage{},
						&types.ContentBlockMemberText{Value: " world"},
					},
				},
			},
			expected: "hello world",
		},
		{
			name: "no text blocks",
			output: &types.ConverseOutputMemberMessage{
				Value: types.Message{
					Content: []types.ContentBlock{
						&types.ContentBlockMemberImage{},
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractTextFromConverseOutput(tt.output, "test-model")

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCompletion_ClaudeModel_UsesTemperatureOnly(t *testing.T) {
	mock := &mockConverseClient{
		converseFunc: func(ctx context.Context, input *bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error) {
			return &bedrockruntime.ConverseOutput{
				Output: &types.ConverseOutputMemberMessage{
					Value: types.Message{
						Content: []types.ContentBlock{
							&types.ContentBlockMemberText{Value: "ok"},
						},
					},
				},
			}, nil
		},
	}

	client := &AmazonBedrockConverseClient{
		client:      mock,
		model:       "anthropic.claude-v2",
		temperature: 0.5,
		topP:        0.9,
		maxTokens:   100,
	}

	_, err := client.GetCompletion(context.Background(), "hello")

	assert.NoError(t, err)

	inf := mock.lastInput.InferenceConfig

	assert.NotNil(t, inf.Temperature)
	assert.Nil(t, inf.TopP)
}

func TestGetCompletion_NonClaudeModel_UsesTemperatureAndTopP(t *testing.T) {
	mock := &mockConverseClient{
		converseFunc: func(ctx context.Context, input *bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error) {
			return &bedrockruntime.ConverseOutput{
				Output: &types.ConverseOutputMemberMessage{
					Value: types.Message{
						Content: []types.ContentBlock{
							&types.ContentBlockMemberText{Value: "ok"},
						},
					},
				},
			}, nil
		},
	}

	client := &AmazonBedrockConverseClient{
		client:      mock,
		model:       "amazon.titan-text",
		temperature: 0.5,
		topP:        0.9,
		maxTokens:   100,
	}

	_, err := client.GetCompletion(context.Background(), "hello")

	assert.NoError(t, err)

	inf := mock.lastInput.InferenceConfig

	assert.NotNil(t, inf.Temperature)
	assert.NotNil(t, inf.TopP)
	assert.Equal(t, float32(0.9), *inf.TopP)
}

func TestIsClaudeModel(t *testing.T) {
	tests := []struct {
		model    string
		expected bool
	}{
		{"anthropic.claude-opus-4-6-v1", true},
		{"CLAUDE-3", true},
		{"amazon.titan", false},
		{"gpt-4", false},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			assert.Equal(t, tt.expected, isClaudeModel(tt.model))
		})
	}
}

func TestGetName(t *testing.T) {
	client := &AmazonBedrockConverseClient{}
	assert.Equal(t, "amazonbedrockconverse", client.GetName())
}
