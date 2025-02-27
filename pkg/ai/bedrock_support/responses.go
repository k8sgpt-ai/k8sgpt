package bedrock_support

import (
	"encoding/json"
	"fmt"
)

type IResponse interface {
	ParseResponse(rawResponse []byte) (string, error)
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

	fmt.Printf("Nova Response\n")

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

	// Assuming you want to return the text content of the message
	if len(response.Output.Message.Content) > 0 {
		return response.Output.Message.Content[0].Text, nil
	}

	return "", nil
}
