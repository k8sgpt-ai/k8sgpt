package bedrock_support

import (
	"encoding/json"
)

type IResponse interface {
	ParseResponse(rawResponse []byte) (string, error)
}

type CohereMessagesResponse struct {
	response IResponse
}

func (a *CohereMessagesResponse) ParseResponse(rawResponse []byte) (string, error) {
	type InvokeModelResponseBody struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Role    string `json:"role"`
		Model   string `json:"model"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StopReason   string      `json:"stop_reason"`
		StopSequence interface{} `json:"stop_sequence"` // Could be null
		Usage        struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	output := &InvokeModelResponseBody{}
	err := json.Unmarshal(rawResponse, output)
	if err != nil {
		return "", err
	}

	// Extract the text content from the Content array
	var resultText string
	for _, content := range output.Content {
		if content.Type == "text" {
			resultText += content.Text
		}
	}

	return resultText, nil
}

type CohereResponse struct {
	response IResponse
}

func (a *CohereResponse) ParseResponse(rawResponse []byte) (string, error) {
	type InvokeModelResponseBody struct {
		Completion  string `json:"completion"`
		Stop_reason string `json:"stop_reason"`
	}
	output := &InvokeModelResponseBody{}
	err := json.Unmarshal(rawResponse, output)
	if err != nil {
		return "", err
	}
	return output.Completion, nil
}

type AI21Response struct {
	response IResponse
}

func (a *AI21Response) ParseResponse(rawResponse []byte) (string, error) {
	type Data struct {
		Text string `json:"text"`
	}
	type Completion struct {
		Data Data `json:"data"`
	}
	type InvokeModelResponseBody struct {
		Completions []Completion `json:"completions"`
	}
	output := &InvokeModelResponseBody{}
	err := json.Unmarshal(rawResponse, output)
	if err != nil {
		return "", err
	}
	return output.Completions[0].Data.Text, nil
}

type AmazonResponse struct {
	response IResponse
}

type NovaResponse struct {
	response NResponse
}
type NResponse interface {
	ParseResponse(rawResponse []byte) (string, error)
}

func (a *AmazonResponse) ParseResponse(rawResponse []byte) (string, error) {
	type Result struct {
		TokenCount       int    `json:"tokenCount"`
		OutputText       string `json:"outputText"`
		CompletionReason string `json:"completionReason"`
	}
	type InvokeModelResponseBody struct {
		InputTextTokenCount int      `json:"inputTextTokenCount"`
		Results             []Result `json:"results"`
	}
	output := &InvokeModelResponseBody{}
	err := json.Unmarshal(rawResponse, output)
	if err != nil {
		return "", err
	}
	return output.Results[0].OutputText, nil
}

func (a *NovaResponse) ParseResponse(rawResponse []byte) (string, error) {
	type Content struct {
		Text string `json:"text"`
	}

	type Message struct {
		Role    string    `json:"role"`
		Content []Content `json:"content"`
	}

	type UsageDetails struct {
		InputTokens               int `json:"inputTokens"`
		OutputTokens              int `json:"outputTokens"`
		TotalTokens               int `json:"totalTokens"`
		CacheReadInputTokenCount  int `json:"cacheReadInputTokenCount"`
		CacheWriteInputTokenCount int `json:"cacheWriteInputTokenCount,omitempty"`
	}

	type AmazonNovaResponse struct {
		Output struct {
			Message Message `json:"message"`
		} `json:"output"`
		StopReason string       `json:"stopReason"`
		Usage      UsageDetails `json:"usage"`
	}

	response := &AmazonNovaResponse{}
	err := json.Unmarshal(rawResponse, response)
	if err != nil {
		return "", err
	}

	if len(response.Output.Message.Content) > 0 {
		return response.Output.Message.Content[0].Text, nil
	}

	return "", nil
}
