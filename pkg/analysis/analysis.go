package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/viper"
)

type Analysis struct {
	Context   context.Context
	Filters   []string
	Client    *kubernetes.Client
	AIClient  ai.IAI
	Results   []analyzer.Result
	Namespace string
	NoCache   bool
	Explain   bool
}

type AnalysisStatus string

const (
	StateOK              AnalysisStatus = "OK"
	StateProblemDetected AnalysisStatus = "ProblemDetected"
)

type JsonOutput struct {
	Status   AnalysisStatus    `json:"status"`
	Problems int               `json:"problems"`
	Results  []analyzer.Result `json:"results"`
}

func (a *Analysis) RunAnalysis() error {

	activeFilters := viper.GetStringSlice("active_filters")

	analyzerMap := analyzer.GetAnalyzerMap()

	analyzerConfig := analyzer.Analyzer{
		Client:    a.Client,
		Context:   a.Context,
		Namespace: a.Namespace,
		AIClient:  a.AIClient,
	}

	// if there are no filters selected and no active_filters then run all of them
	if len(a.Filters) == 0 && len(activeFilters) == 0 {
		for _, analyzer := range analyzerMap {
			results, err := analyzer.Analyze(analyzerConfig)
			if err != nil {
				return err
			}
			a.Results = append(a.Results, results...)
		}
		return nil
	}

	// if the filters flag is specified
	if len(a.Filters) != 0 {
		for _, filter := range a.Filters {
			if analyzer, ok := analyzerMap[filter]; ok {
				results, err := analyzer.Analyze(analyzerConfig)
				if err != nil {
					return err
				}
				a.Results = append(a.Results, results...)
			}
		}
		return nil
	}

	// use active_filters
	for _, filter := range activeFilters {
		if analyzer, ok := analyzerMap[filter]; ok {
			results, err := analyzer.Analyze(analyzerConfig)
			if err != nil {
				return err
			}
			a.Results = append(a.Results, results...)
		}
	}
	return nil
}

func (a *Analysis) JsonOutput() ([]byte, error) {
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
		Status:   status,
	}
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error marshalling json: %v", err)
	}
	return output, nil
}

func (a *Analysis) PrintOutput() {
	fmt.Println("")
	if len(a.Results) == 0 {
		fmt.Println(color.GreenString("No problems detected"))
	}
	for n, result := range a.Results {
		fmt.Printf("%s %s(%s)\n", color.CyanString("%d", n),
			color.YellowString(result.Name), color.CyanString(result.ParentObject))
		for _, err := range result.Error {
			fmt.Printf("- %s %s\n", color.RedString("Error:"), color.RedString(err.Text))
		}
		fmt.Println(color.GreenString(result.Details + "\n"))
	}
}

func (a *Analysis) GetAIResults(output string, anonymize bool) error {
	if len(a.Results) == 0 {
		return nil
	}

	var bar *progressbar.ProgressBar
	if output != "json" {
		bar = progressbar.Default(int64(len(a.Results)))
	}

	for index, analysis := range a.Results {
		var texts []string

		for _, failure := range analysis.Error {
			if anonymize {
				for _, s := range failure.Sensitive {
					failure.Text = util.ReplaceIfMatch(failure.Text, s.Unmasked, s.Masked)
				}
			}
			texts = append(texts, failure.Text)
		}
		parsedText, err := a.AIClient.Parse(a.Context, texts, a.NoCache)
		if err != nil {
			// FIXME: can we avoid checking if output is json multiple times?
			//   maybe implement the progress bar better?
			if output != "json" {
				bar.Exit()
			}

			// Check for exhaustion
			if strings.Contains(err.Error(), "status code: 429") {
				return fmt.Errorf("exhausted API quota for AI provider %s: %v", a.AIClient.GetName(), err)
			} else {
				return fmt.Errorf("failed while calling AI provider %s: %v", a.AIClient.GetName(), err)
			}
		}

		if anonymize {
			for _, failure := range analysis.Error {
				for _, s := range failure.Sensitive {
					parsedText = strings.ReplaceAll(parsedText, s.Masked, s.Unmasked)
				}
			}
		}

		analysis.Details = parsedText
		if output != "json" {
			bar.Add(1)
		}
		a.Results[index] = analysis
	}
	return nil
}
