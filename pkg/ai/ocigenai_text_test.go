package ai

import (
	"testing"

	"github.com/oracle/oci-go-sdk/v65/generativeaiinference"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractGeneratedText_CohereNilText(t *testing.T) {
	out, err := extractGeneratedText(generativeaiinference.CohereChatResponse{
		FinishReason: generativeaiinference.CohereChatResponseFinishReasonError,
	})

	require.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, err.Error(), "no text found")
}

func TestExtractGeneratedText_CohereBlankText(t *testing.T) {
	blank := "  "
	out, err := extractGeneratedText(generativeaiinference.CohereChatResponse{
		Text:         &blank,
		FinishReason: generativeaiinference.CohereChatResponseFinishReasonMaxTokens,
	})

	require.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, err.Error(), "empty completion")
	assert.Contains(t, err.Error(), "MAX_TOKENS")
}

func TestExtractGeneratedText_CohereText(t *testing.T) {
	text := "the pod is crash looping"
	out, err := extractGeneratedText(generativeaiinference.CohereChatResponse{
		Text:         &text,
		FinishReason: generativeaiinference.CohereChatResponseFinishReasonComplete,
	})

	require.NoError(t, err)
	assert.Equal(t, text, out)
}

func TestExtractGeneratedText_GenericNilText(t *testing.T) {
	out, err := extractGeneratedText(generativeaiinference.GenericChatResponse{
		Choices: []generativeaiinference.ChatChoice{{
			Message: generativeaiinference.AssistantMessage{
				Content: []generativeaiinference.ChatContent{
					generativeaiinference.TextContent{},
				},
			},
		}},
	})

	require.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, err.Error(), "no text found")
}

func TestExtractGeneratedText_GenericBlankText(t *testing.T) {
	blank := ""
	finish := "length"
	out, err := extractGeneratedText(generativeaiinference.GenericChatResponse{
		Choices: []generativeaiinference.ChatChoice{{
			FinishReason: &finish,
			Message: generativeaiinference.AssistantMessage{
				Content: []generativeaiinference.ChatContent{
					generativeaiinference.TextContent{Text: &blank},
				},
			},
		}},
	})

	require.Error(t, err)
	assert.Empty(t, out)
	assert.Contains(t, err.Error(), "empty completion")
	assert.Contains(t, err.Error(), "finish reason: length")
}

func TestExtractGeneratedText_GenericText(t *testing.T) {
	text := "ok"
	out, err := extractGeneratedText(generativeaiinference.GenericChatResponse{
		Choices: []generativeaiinference.ChatChoice{{
			Message: generativeaiinference.AssistantMessage{
				Content: []generativeaiinference.ChatContent{
					generativeaiinference.TextContent{Text: &text},
				},
			},
		}},
	})

	require.NoError(t, err)
	assert.Equal(t, text, out)
}
