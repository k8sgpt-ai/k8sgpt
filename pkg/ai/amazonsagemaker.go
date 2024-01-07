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
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sagemakerruntime"
)

const amazonsagemakerAIClientName = "amazonsagemaker"

type SageMakerAIClient struct {
	nopCloser

	client      *sagemakerruntime.SageMakerRuntime
	model       string
	temperature float32
	endpoint    string
	topP        float32
	maxTokens   int
}

type Generations []struct {
	Generation struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"generation"`
}

type Request struct {
	Inputs     [][]Message `json:"inputs"`
	Parameters Parameters  `json:"parameters"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Parameters struct {
	MaxNewTokens int     `json:"max_new_tokens"`
	TopP         float64 `json:"top_p"`
	Temperature  float64 `json:"temperature"`
}

func (c *SageMakerAIClient) Configure(config IAIConfig) error {

	// Create a new AWS session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(config.GetProviderRegion())},
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create a new SageMaker runtime client
	c.client = sagemakerruntime.New(sess)
	c.model = config.GetModel()
	c.endpoint = config.GetEndpointName()
	c.temperature = config.GetTemperature()
	c.maxTokens = config.GetMaxTokens()
	c.topP = config.GetTopP()
	return nil
}

func (c *SageMakerAIClient) GetCompletion(_ context.Context, prompt string) (string, error) {
	// Create a completion request
	request := Request{
		Inputs: [][]Message{
			{
				{Role: "system", Content: "DEFAULT_PROMPT"},
				{Role: "user", Content: prompt},
			},
		},

		Parameters: Parameters{
			MaxNewTokens: int(c.maxTokens),
			TopP:         float64(c.topP),
			Temperature:  float64(c.temperature),
		},
	}

	// Convert request to []byte
	bytesData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	// Create an input object
	input := &sagemakerruntime.InvokeEndpointInput{
		Body:             bytesData,
		EndpointName:     aws.String(c.endpoint),
		ContentType:      aws.String("application/json"), // Set the content type as per your model's requirements
		Accept:           aws.String("application/json"), // Set the accept type as per your model's requirements
		CustomAttributes: aws.String("accept_eula=true"),
	}

	// Call the InvokeEndpoint function
	result, err := c.client.InvokeEndpoint(input)
	if err != nil {
		return "", err
	}

	// // Define a slice of Generations
	var generations Generations

	err = json.Unmarshal([]byte(string(result.Body)), &generations)
	if err != nil {
		return "", err
	}
	// Check for length of generations
	if len(generations) != 1 {
		return "", fmt.Errorf("Expected exactly one generation, but got %d", len(generations))
	}

	// Access the content
	content := generations[0].Generation.Content
	return content, nil
}

func (c *SageMakerAIClient) GetName() string {
	return amazonsagemakerAIClientName
}
