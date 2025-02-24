package bedrock_support

import "encoding/json"

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
