package ai

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildOpenAIResponsesURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    string
	}{
		{
			name:    "default",
			baseURL: "",
			want:    defaultOpenAIResponsesURL,
		},
		{
			name:    "appendResponses",
			baseURL: "https://api.openai.com/v1",
			want:    "https://api.openai.com/v1/responses",
		},
		{
			name:    "appendResponsesTrailingSlash",
			baseURL: "https://api.openai.com/v1/",
			want:    "https://api.openai.com/v1/responses",
		},
		{
			name:    "alreadyResponses",
			baseURL: "https://api.openai.com/v1/responses",
			want:    "https://api.openai.com/v1/responses",
		},
		{
			name:    "customHost",
			baseURL: "http://localhost:8080/v1",
			want:    "http://localhost:8080/v1/responses",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := buildOpenAIResponsesURL(test.baseURL)
			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
}

func TestExtractOpenAIResponsesText(t *testing.T) {
	response := openAIResponsesResponse{
		Output: []openAIResponsesOutputItem{
			{
				Type: "message",
				Content: []openAIResponsesTextChunk{
					{Type: "output_text", Text: "Hello"},
					{Type: "output_text", Text: " world"},
				},
			},
			{
				Type: "message",
				Content: []openAIResponsesTextChunk{
					{Type: "output_text", Text: "!"},
				},
			},
			{
				Type: "tool_call",
			},
		},
	}

	got := extractOpenAIResponsesText(response)
	require.Equal(t, "Hello world!", got)
}

func TestExtractOpenAIResponsesTextReasoning(t *testing.T) {
	response := openAIResponsesResponse{
		Output: []openAIResponsesOutputItem{
			{
				Type: "message",
				Content: []openAIResponsesTextChunk{
					{Type: "reasoning_text", Text: "Reasoning "},
					{Type: "summary_text", Text: "summary "},
					{Type: "refusal", Refusal: "refusal"},
				},
			},
		},
	}

	got := extractOpenAIResponsesText(response)
	require.Equal(t, "Reasoning summary refusal", got)
}

func TestExtractOpenAIResponsesTextFallback(t *testing.T) {
	response := openAIResponsesResponse{
		OutputText: "fallback output",
	}

	got := extractOpenAIResponsesText(response)
	require.Equal(t, "fallback output", got)
}

func TestSupportsOpenAIResponsesSamplingParams(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  bool
	}{
		{name: "gpt-4o", model: "gpt-4o", want: true},
		{name: "gpt-4.1", model: "gpt-4.1", want: true},
		{name: "gpt-5-nano", model: "gpt-5-nano", want: false},
		{name: "o1", model: "o1", want: false},
		{name: "o3-mini", model: "o3-mini", want: false},
		{name: "o4-mini", model: "o4-mini", want: false},
		{name: "empty", model: "", want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, supportsOpenAIResponsesSamplingParams(test.model))
		})
	}
}

func TestRequiresOpenAIResponsesReasoningConfig(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  bool
	}{
		{name: "gpt-4o", model: "gpt-4o", want: false},
		{name: "gpt-5-nano", model: "gpt-5-nano", want: true},
		{name: "o1", model: "o1", want: true},
		{name: "o3-mini", model: "o3-mini", want: true},
		{name: "empty", model: "", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, requiresOpenAIResponsesReasoningConfig(test.model))
		})
	}
}
