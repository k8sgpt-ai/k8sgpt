// Copyright © 2023 K8sGPT.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package analysis

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

var outputFormats = map[string]func(*Analysis) ([]byte, error){
	"json": (*Analysis).jsonOutput,
	"text": (*Analysis).textOutput,
}

func getOutputFormats() []string {
	formats := make([]string, 0, len(outputFormats))
	for format := range outputFormats {
		formats = append(formats, format)
	}
	return formats
}

func (a *Analysis) PrintOutput(format string) ([]byte, error) {
	outputFunc, ok := outputFormats[format]
	if !ok {
		return nil, fmt.Errorf("unsupported output format: %s. Available format %s", format, strings.Join(getOutputFormats(), ","))
	}
	return outputFunc(a)
}

func (a *Analysis) jsonOutput() ([]byte, error) {
	var problems int
	var status AnalysisStatus
	for _, result := range a.Results {
		problems += len(result.Error)
	}
	if problems > 0 {
		status = StateProblemDetected
	} else {
		status = StateOK
	}

	result := JsonOutput{
		Problems: problems,
		Results:  a.Results,
		Errors:   a.Errors,
		Status:   status,
	}
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshaling json: %v", err)
	}
	return output, nil
}

func (a *Analysis) textOutput() ([]byte, error) {
	var output strings.Builder
	if len(a.Errors) != 0 {
		output.WriteString("\n")
		output.WriteString(color.YellowString("Warnings : \n"))
		for _, aerror := range a.Errors {
			output.WriteString(fmt.Sprintf("- %s\n", color.YellowString(aerror)))
		}
	}
	output.WriteString("\n")
	if len(a.Results) == 0 {
		output.WriteString(color.GreenString("No problems detected\n"))
		return []byte(output.String()), nil
	}
	for n, result := range a.Results {
		output.WriteString(fmt.Sprintf("%s %s(%s)\n", color.CyanString("%d", n),
			color.YellowString(result.Name), color.CyanString(result.ParentObject)))
		for _, err := range result.Error {
			output.WriteString(fmt.Sprintf("- %s %s\n", color.RedString("Error:"), color.RedString(err.Text)))
		}
		output.WriteString(color.GreenString(result.Details + "\n"))
	}
	return []byte(output.String()), nil
}
