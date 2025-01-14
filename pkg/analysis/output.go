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
		Provider: a.AnalysisAIProvider,
		Problems: problems,
		Results:  a.Results,
		Errors:   a.Errors,
		Status:   status,
	}
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshalling json: %v", err)
	}
	return output, nil
}

func (a *Analysis) PrintStats() []byte {
	var output strings.Builder

	output.WriteString(color.YellowString("The stats mode allows for debugging and understanding the time taken by an analysis by displaying the statistics of each analyzer.\n"))

	for _, stat := range a.Stats {
		output.WriteString(fmt.Sprintf("- Analyzer %s took %s \n", color.YellowString(stat.Analyzer), stat.DurationTime))
	}

	return []byte(output.String())
}

func (a *Analysis) textOutput() ([]byte, error) {
	var output strings.Builder

	// Print the AI provider used for this analysis (if explain was enabled).
	if a.Explain {
		output.WriteString(fmt.Sprintf("AI Provider: %s\n", color.YellowString(a.AnalysisAIProvider)))
	} else {
		output.WriteString(fmt.Sprintf("AI Provider: %s\n", color.YellowString("AI not used; --explain not set")))
	}

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
		output.WriteString(fmt.Sprintf("%s: %s %s(%s)\n", color.CyanString("%d", n),
			color.HiYellowString(result.Kind),
			color.YellowString(result.Name),
			color.CyanString(result.ParentObject)))
		for _, err := range result.Error {
			output.WriteString(fmt.Sprintf("- %s %s\n", color.RedString("Error:"), color.RedString(err.Text)))
			if err.KubernetesDoc != "" {
				output.WriteString(fmt.Sprintf("  %s %s\n", color.RedString("Kubernetes Doc:"), color.RedString(err.KubernetesDoc)))
			}
		}
		output.WriteString(color.GreenString(result.Details + "\n"))
	}
	return []byte(output.String()), nil
}
