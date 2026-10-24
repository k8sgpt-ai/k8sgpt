package analysis

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
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

// resultsWithFindingHashes returns the analysis results with Failure.Hash
// filled in for every failure. The hash is derived from the same inputs
// ApplyExclusions matches on (Kind, Name and the pre-anonymization failure
// text), so the value printed here is the one a user pastes into the
// exclusions config. The receiver's results are left untouched: this is a
// rendering-only projection, which keeps the hash out of the AI payload and
// out of the exclusion matching itself.
func (a *Analysis) resultsWithFindingHashes() []common.Result {
	if len(a.Results) == 0 {
		return a.Results
	}
	results := make([]common.Result, len(a.Results))
	for i, result := range a.Results {
		results[i] = result
		if len(result.Error) == 0 {
			continue
		}
		failures := make([]common.Failure, len(result.Error))
		for j, failure := range result.Error {
			failure.Hash = FindingHash(result.Kind, result.Name, failure.Text)
			failures[j] = failure
		}
		results[i].Error = failures
	}
	return results
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
		Excluded: a.Excluded,
		Results:  a.resultsWithFindingHashes(),
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
		if a.Excluded > 0 {
			output.WriteString(color.GreenString("%d findings excluded\n", a.Excluded))
		}
		return []byte(output.String()), nil
	}
	for n, result := range a.Results {
		output.WriteString(fmt.Sprintf("%s: %s %s(%s)\n", color.CyanString("%d", n),
			color.HiYellowString(result.Kind),
			color.YellowString(result.Name),
			color.CyanString(result.ParentObject)))
		for _, err := range result.Error {
			// The trailing hash is what a user copies into the exclusions
			// config; without it the feature is unusable without recomputing
			// SHA-256 by hand.
			output.WriteString(fmt.Sprintf("- %s %s %s\n", color.RedString("Error:"), color.RedString(err.Text),
				color.HiBlackString("hash=%s", FindingHash(result.Kind, result.Name, err.Text))))
			if err.KubernetesDoc != "" {
				output.WriteString(fmt.Sprintf("  %s %s\n", color.RedString("Kubernetes Doc:"), color.RedString(err.KubernetesDoc)))
			}
		}
		output.WriteString(color.GreenString(result.Details + "\n"))
	}
	if a.Excluded > 0 {
		output.WriteString(color.YellowString("\n%d findings excluded\n", a.Excluded))
	}
	return []byte(output.String()), nil
}
