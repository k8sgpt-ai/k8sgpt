package query

import (
	"context"
	"errors"
	"testing"

	schemav1 "buf.build/gen/go/k8sgpt-ai/k8sgpt/protocolbuffers/go/schema/v1"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAI is a mock implementation of the ai.IAI interface for testing
type MockAI struct {
	mock.Mock
}

func (m *MockAI) Configure(config ai.IAIConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockAI) GetCompletion(ctx context.Context, prompt string) (string, error) {
	args := m.Called(ctx, prompt)
	return args.String(0), args.Error(1)
}

func (m *MockAI) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAI) Close() {
	m.Called()
}

// MockAIClientFactory is a mock implementation of AIClientFactory
type MockAIClientFactory struct {
	mock.Mock
}

func (m *MockAIClientFactory) NewClient(provider string) ai.IAI {
	args := m.Called(provider)
	return args.Get(0).(ai.IAI)
}

// MockConfigProvider is a mock implementation of ConfigProvider
type MockConfigProvider struct {
	mock.Mock
}

func (m *MockConfigProvider) UnmarshalKey(key string, rawVal interface{}) error {
	args := m.Called(key, rawVal)

	// If we want to set the rawVal (which is a pointer)
	if fn, ok := args.Get(0).(func(interface{})); ok && fn != nil {
		fn(rawVal)
	}

	// Return the error as the first return value
	return args.Error(0)
}

func TestQuery_Success(t *testing.T) {
	// Setup mocks
	mockAI := new(MockAI)
	mockFactory := new(MockAIClientFactory)
	mockConfig := new(MockConfigProvider)

	// Set test implementations
	ai.SetTestAIClientFactory(mockFactory)
	ai.SetTestConfigProvider(mockConfig)
	defer ai.ResetTestImplementations()

	// Define test data
	testBackend := "test-backend"
	testQuery := "test query"
	testResponse := "test response"

	// Setup expectations
	mockFactory.On("NewClient", testBackend).Return(mockAI)
	mockAI.On("Close").Return()

	// Set up configuration with a valid provider
	mockConfig.On("UnmarshalKey", "ai", mock.Anything).Run(func(args mock.Arguments) {
		config := args.Get(1).(*ai.AIConfiguration)
		*config = ai.AIConfiguration{
			Providers: []ai.AIProvider{
				{
					Name:     testBackend,
					Password: "test-password",
					Model:    "test-model",
				},
			},
		}
	}).Return(nil)

	mockAI.On("Configure", mock.AnythingOfType("*ai.AIProvider")).Return(nil)
	mockAI.On("GetCompletion", mock.Anything, testQuery).Return(testResponse, nil)

	// Create handler and call Query
	handler := &Handler{}
	response, err := handler.Query(context.Background(), &schemav1.QueryRequest{
		Backend: testBackend,
		Query:   testQuery,
	})

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, testResponse, response.Response)
	assert.Equal(t, "", response.Error.Message)

	// Verify mocks
	mockAI.AssertExpectations(t)
	mockFactory.AssertExpectations(t)
	mockConfig.AssertExpectations(t)
}

func TestQuery_UnmarshalError(t *testing.T) {
	// Setup mocks
	mockAI := new(MockAI)
	mockFactory := new(MockAIClientFactory)
	mockConfig := new(MockConfigProvider)

	// Set test implementations
	ai.SetTestAIClientFactory(mockFactory)
	ai.SetTestConfigProvider(mockConfig)
	defer ai.ResetTestImplementations()

	// Setup expectations
	mockFactory.On("NewClient", "test-backend").Return(mockAI)
	mockAI.On("Close").Return()

	// Mock unmarshal error
	mockConfig.On("UnmarshalKey", "ai", mock.Anything).Return(errors.New("unmarshal error"))

	// Create handler and call Query
	handler := &Handler{}
	response, err := handler.Query(context.Background(), &schemav1.QueryRequest{
		Backend: "test-backend",
		Query:   "test query",
	})

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "", response.Response)
	assert.Contains(t, response.Error.Message, "Failed to unmarshal AI configuration")

	// Verify mocks
	mockAI.AssertExpectations(t)
	mockFactory.AssertExpectations(t)
	mockConfig.AssertExpectations(t)
}

func TestQuery_ProviderNotFound(t *testing.T) {
	// Setup mocks
	mockAI := new(MockAI)
	mockFactory := new(MockAIClientFactory)
	mockConfig := new(MockConfigProvider)

	// Set test implementations
	ai.SetTestAIClientFactory(mockFactory)
	ai.SetTestConfigProvider(mockConfig)
	defer ai.ResetTestImplementations()

	// Define test data
	testBackend := "test-backend"

	// Setup expectations
	mockFactory.On("NewClient", testBackend).Return(mockAI)
	mockAI.On("Close").Return()

	// Set up configuration with no matching provider
	mockConfig.On("UnmarshalKey", "ai", mock.Anything).Run(func(args mock.Arguments) {
		config := args.Get(1).(*ai.AIConfiguration)
		*config = ai.AIConfiguration{
			Providers: []ai.AIProvider{
				{
					Name: "other-backend",
				},
			},
		}
	}).Return(nil)

	// Create handler and call Query
	handler := &Handler{}
	response, err := handler.Query(context.Background(), &schemav1.QueryRequest{
		Backend: testBackend,
		Query:   "test query",
	})

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "", response.Response)
	assert.Contains(t, response.Error.Message, "AI provider test-backend not found in configuration")

	// Verify mocks
	mockAI.AssertExpectations(t)
	mockFactory.AssertExpectations(t)
	mockConfig.AssertExpectations(t)
}

func TestQuery_ConfigureError(t *testing.T) {
	// Setup mocks
	mockAI := new(MockAI)
	mockFactory := new(MockAIClientFactory)
	mockConfig := new(MockConfigProvider)

	// Set test implementations
	ai.SetTestAIClientFactory(mockFactory)
	ai.SetTestConfigProvider(mockConfig)
	defer ai.ResetTestImplementations()

	// Define test data
	testBackend := "test-backend"

	// Setup expectations
	mockFactory.On("NewClient", testBackend).Return(mockAI)
	mockAI.On("Close").Return()

	// Set up configuration with a valid provider
	mockConfig.On("UnmarshalKey", "ai", mock.Anything).Run(func(args mock.Arguments) {
		config := args.Get(1).(*ai.AIConfiguration)
		*config = ai.AIConfiguration{
			Providers: []ai.AIProvider{
				{
					Name: testBackend,
				},
			},
		}
	}).Return(nil)

	// Mock configure error
	mockAI.On("Configure", mock.AnythingOfType("*ai.AIProvider")).Return(errors.New("configure error"))

	// Create handler and call Query
	handler := &Handler{}
	response, err := handler.Query(context.Background(), &schemav1.QueryRequest{
		Backend: testBackend,
		Query:   "test query",
	})

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "", response.Response)
	assert.Contains(t, response.Error.Message, "Failed to configure AI client")

	// Verify mocks
	mockAI.AssertExpectations(t)
	mockFactory.AssertExpectations(t)
	mockConfig.AssertExpectations(t)
}

func TestQuery_GetCompletionError(t *testing.T) {
	// Setup mocks
	mockAI := new(MockAI)
	mockFactory := new(MockAIClientFactory)
	mockConfig := new(MockConfigProvider)

	// Set test implementations
	ai.SetTestAIClientFactory(mockFactory)
	ai.SetTestConfigProvider(mockConfig)
	defer ai.ResetTestImplementations()

	// Define test data
	testBackend := "test-backend"
	testQuery := "test query"

	// Setup expectations
	mockFactory.On("NewClient", testBackend).Return(mockAI)
	mockAI.On("Close").Return()

	// Set up configuration with a valid provider
	mockConfig.On("UnmarshalKey", "ai", mock.Anything).Run(func(args mock.Arguments) {
		config := args.Get(1).(*ai.AIConfiguration)
		*config = ai.AIConfiguration{
			Providers: []ai.AIProvider{
				{
					Name: testBackend,
				},
			},
		}
	}).Return(nil)

	mockAI.On("Configure", mock.AnythingOfType("*ai.AIProvider")).Return(nil)
	mockAI.On("GetCompletion", mock.Anything, testQuery).Return("", errors.New("completion error"))

	// Create handler and call Query
	handler := &Handler{}
	response, err := handler.Query(context.Background(), &schemav1.QueryRequest{
		Backend: testBackend,
		Query:   testQuery,
	})

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "", response.Response)
	assert.Equal(t, "completion error", response.Error.Message)

	// Verify mocks
	mockAI.AssertExpectations(t)
	mockFactory.AssertExpectations(t)
	mockConfig.AssertExpectations(t)
}