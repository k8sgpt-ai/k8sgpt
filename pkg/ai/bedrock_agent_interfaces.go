package ai

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
)

// BedrockAgentAPI defines the interface for Bedrock Agent operations
// This interface is used for Knowledge Base operations
type BedrockAgentAPI interface {
	// GetKnowledgeBase retrieves information about a knowledge base
	GetKnowledgeBase(ctx context.Context, params *bedrockagent.GetKnowledgeBaseInput, optFns ...func(*bedrockagent.Options)) (*bedrockagent.GetKnowledgeBaseOutput, error)
}

// BedrockAgentRuntimeAPI defines the interface for Bedrock Agent Runtime operations
// This interface is used for Knowledge Base retrieval and generation
type BedrockAgentRuntimeAPI interface {
	// RetrieveAndGenerate performs retrieval-augmented generation using knowledge bases
	RetrieveAndGenerate(ctx context.Context, params *bedrockagentruntime.RetrieveAndGenerateInput, optFns ...func(*bedrockagentruntime.Options)) (*bedrockagentruntime.RetrieveAndGenerateOutput, error)
}
