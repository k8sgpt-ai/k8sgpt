/*
Copyright 2023 The K8sGPT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime/types"
)

const amazonbedrockKnowledgeBaseAIClientName = "amazonbedrockknowledgebase"

// KnowledgeBaseConfig represents the configuration for a knowledge base
type KnowledgeBaseConfig struct {
	ID              string `json:"id"`
	NumberOfResults int32  `json:"numberOfResults"`
}

// AmazonBedRockKnowledgeBaseClient represents the client for interacting with the Amazon Bedrock Knowledge Base service.
type AmazonBedRockKnowledgeBaseClient struct {
	nopCloser

	agentClient         BedrockAgentAPI
	agentRuntimeClient  BedrockAgentRuntimeAPI
	knowledgeBases      []KnowledgeBaseConfig
	modelId             string
	temperature         float32
	topP                float32
	maxTokens           int
	enableCitations     bool
}

// NewAmazonBedRockKnowledgeBaseClient creates a new AmazonBedRockKnowledgeBaseClient
func NewAmazonBedRockKnowledgeBaseClient() *AmazonBedRockKnowledgeBaseClient {
	return &AmazonBedRockKnowledgeBaseClient{
		enableCitations: true, // Enable citations by default
	}
}

// Configure configures the AmazonBedRockKnowledgeBaseClient with the provided configuration.
func (a *AmazonBedRockKnowledgeBaseClient) Configure(config IAIConfig) error {
	// Get the knowledge base ID from the KnowledgeBase field
	knowledgeBase := config.GetKnowledgeBase()
	if knowledgeBase == "" {
		return errors.New("knowledge base is required for Amazon Bedrock Knowledge Base integration")
	}

	// Create a single knowledge base configuration
	a.knowledgeBases = []KnowledgeBaseConfig{
		{
			ID:              knowledgeBase,
			NumberOfResults: 5, // Default number of results
		},
	}

	// Get the model ID
	modelId := config.GetModel()
	if modelId == "" {
		return errors.New("model ID is required for Amazon Bedrock Knowledge Base integration")
	}
	a.modelId = modelId

	// Determine the appropriate region to use
	region := GetRegionOrDefault(config.GetProviderRegion())

	// Only create AWS clients if they haven't been injected (for testing)
	if a.agentClient == nil || a.agentRuntimeClient == nil {
		// Create a new AWS config with the determined region
		cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion(region),
		)
		if err != nil {
			return fmt.Errorf("failed to load AWS config for region %s: %w", region, err)
		}

		// Create clients with the config
		a.agentClient = bedrockagent.NewFromConfig(cfg)
		a.agentRuntimeClient = bedrockagentruntime.NewFromConfig(cfg)
	}

	// Validate that all knowledge bases exist
	for _, kb := range a.knowledgeBases {
		_, err := a.agentClient.GetKnowledgeBase(context.Background(), &bedrockagent.GetKnowledgeBaseInput{
			KnowledgeBaseId: aws.String(kb.ID),
		})
		if err != nil {
			return fmt.Errorf("failed to get knowledge base %s: %w", kb.ID, err)
		}
	}

	// Set common configuration parameters
	a.temperature = config.GetTemperature()
	a.topP = config.GetTopP()
	a.maxTokens = config.GetMaxTokens()

	return nil
}

// getUniqueSnippets returns a slice of unique snippets, up to the specified maximum number
func getUniqueSnippets(snippets []string, maxSnippets int) []string {
	if len(snippets) == 0 {
		return []string{}
	}

	// Use a map to track unique prefixes we've seen
	seen := make(map[string]bool)
	result := []string{}

	for _, snippet := range snippets {
		// Get a prefix that's representative of the snippet
		prefix := ""
		if len(snippet) > 30 {
			prefix = snippet[:30]
		} else {
			prefix = snippet
		}

		// If we haven't seen this prefix before, add it
		if !seen[prefix] {
			seen[prefix] = true
			
			// Truncate snippet if needed
			displaySnippet := snippet
			if len(displaySnippet) > 50 {
				displaySnippet = displaySnippet[:50] + "..."
			}
			
			result = append(result, displaySnippet)
			
			// Stop if we've reached the maximum number of snippets
			if len(result) >= maxSnippets {
				break
			}
		}
	}

	return result
}

// GetCompletion sends a request to the model for generating completion based on the provided prompt.
func (a *AmazonBedRockKnowledgeBaseClient) GetCompletion(ctx context.Context, prompt string) (string, error) {
	// Create the knowledge base configuration
	kbConfig := &types.KnowledgeBaseRetrieveAndGenerateConfiguration{
		KnowledgeBaseId: aws.String(a.knowledgeBases[0].ID), // Use the single knowledge base
		ModelArn:        aws.String(a.modelId),
	}

	// Add retrieval configuration
	kbConfig.RetrievalConfiguration = &types.KnowledgeBaseRetrievalConfiguration{
		VectorSearchConfiguration: &types.KnowledgeBaseVectorSearchConfiguration{
			NumberOfResults: aws.Int32(a.knowledgeBases[0].NumberOfResults),
		},
	}

	// Add inference configuration with text inference settings
	kbConfig.GenerationConfiguration = &types.GenerationConfiguration{
		InferenceConfig: &types.InferenceConfig{
			TextInferenceConfig: &types.TextInferenceConfig{
				Temperature: aws.Float32(a.temperature),
				TopP:        aws.Float32(a.topP),
				MaxTokens:   aws.Int32(int32(a.maxTokens)),
			},
		},
	}

	// Configure citations if enabled
	if a.enableCitations {
		// Create a prompt template similar to the default_prompt but adapted for Knowledge Base
		citationPrompt := "Simplify the following Kubernetes error message using the retrieved information; " +
			"$search_results$\n\n" +
			"Question: $question$\n\n" +
			"Provide the most possible solution in a step by step style. " +
			"Write the output in the following format:\n" +
			"Error: {Explain error here}\n\n" +
			"Solution:\n" +
			"1. {First step}\n\n" +
			"2. {Second step}\n\n" +
			"3. {Third step if needed}\n\n" +
			"4. {Fourth step if needed}\n\n" +
			"Include sources at the end if relevant."

		// Add citation configuration
		kbConfig.GenerationConfiguration.PromptTemplate = &types.PromptTemplate{
			TextPromptTemplate: aws.String(citationPrompt),
		}
	}

	// Create the retrieval and generate input
	input := &bedrockagentruntime.RetrieveAndGenerateInput{
		Input: &types.RetrieveAndGenerateInput{
			Text: aws.String(prompt),
		},
		RetrieveAndGenerateConfiguration: &types.RetrieveAndGenerateConfiguration{
			Type:                       types.RetrieveAndGenerateTypeKnowledgeBase,
			KnowledgeBaseConfiguration: kbConfig,
		},
	}

	// Call the RetrieveAndGenerate API
	response, err := a.agentRuntimeClient.RetrieveAndGenerate(ctx, input)
	if err != nil {
		return "", err
	}

	// Extract the generated text from the response
	if response.Output == nil || response.Output.Text == nil {
		return "", errors.New("no output text in response")
	}

	result := *response.Output.Text
	
	// Process citations and enhance the response
	citationMap := make(map[string][]string) // Map document URI to snippets
	documentOrder := []string{} // Preserve document order
	sourcesSection := "\n\nSources:\n"
	
	// Process citations if available
	if response.Citations != nil && len(response.Citations) > 0 {
		for _, citation := range response.Citations {
			sourceURI := ""
			snippetText := ""
			
			if citation.RetrievedReferences != nil && len(citation.RetrievedReferences) > 0 {
				for _, ref := range citation.RetrievedReferences {
					if ref.Location != nil && ref.Location.S3Location != nil && ref.Location.S3Location.Uri != nil {
						sourceURI = *ref.Location.S3Location.Uri
					}
					
					// Get snippet content if available
					if ref.Content != nil && ref.Content.Text != nil {
						snippetText = *ref.Content.Text
					}
				}
			}
			
			// Store citation info in map, grouping by document URI
			if sourceURI != "" {
				// Check if this document is already in our map
				if _, exists := citationMap[sourceURI]; !exists {
					// First time seeing this document, add to order list
					documentOrder = append(documentOrder, sourceURI)
				}
				
				// Add snippet to the document's snippets list
				if snippetText != "" {
					citationMap[sourceURI] = append(citationMap[sourceURI], snippetText)
				}
			}
		}
		
		// Build deduplicated sources section
		if len(documentOrder) > 0 {
			// Check if the model already included a sources section
			if !strings.Contains(strings.ToLower(result), "sources:") {
				// Build sources section with deduplicated documents but showing different snippets
				for i, docURI := range documentOrder {
					snippets := citationMap[docURI]
					docLabel := fmt.Sprintf("[doc%d]", i+1)
					
					// Get the filename from the URI
					parts := strings.Split(docURI, "/")
					filename := docURI
					if len(parts) > 0 {
						filename = parts[len(parts)-1]
					}
					
					// Format snippets to show variety
					snippetDisplay := ""
					if len(snippets) > 0 {
						// Show up to 2 different snippets
						uniqueSnippets := getUniqueSnippets(snippets, 2)
						if len(uniqueSnippets) > 0 {
							snippetDisplay = fmt.Sprintf(" (Snippets: %s)", strings.Join(uniqueSnippets, "; "))
						}
					}
					
					sourcesSection += fmt.Sprintf("%s: %s%s\n", docLabel, filename, snippetDisplay)
				}
				
				result += sourcesSection
			}
		}
	} else if a.enableCitations {
		// No citations were returned but citations are enabled
		result += "\n\nNote: This response was generated using knowledge base information, but specific citations could not be provided."
	}
	
	return result, nil
}

// GetName returns the name of the AmazonBedRockKnowledgeBaseClient.
func (a *AmazonBedRockKnowledgeBaseClient) GetName() string {
	return amazonbedrockKnowledgeBaseAIClientName
}
