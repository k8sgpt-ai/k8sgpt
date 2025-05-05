package ai

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// BedrockManagementAPI defines the interface for Bedrock management operations
type BedrockManagementAPI interface {
	GetInferenceProfile(ctx context.Context, params *bedrock.GetInferenceProfileInput, optFns ...func(*bedrock.Options)) (*bedrock.GetInferenceProfileOutput, error)
}

// BedrockRuntimeAPI defines the interface for Bedrock runtime operations
type BedrockRuntimeAPI interface {
	InvokeModel(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error)
}
