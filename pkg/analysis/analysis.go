package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/viper"
	"os"
	"strings"
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
			fmt.Printf("- %s %s\n", color.RedString("Error:"), color.RedString(err))
		}
		fmt.Println(color.GreenString(result.Details + "\n"))
	}
}

func (a *Analysis) GetAIResults(progressBar bool) error {
	if len(a.Results) == 0 {
		return nil
	}

	var bar *progressbar.ProgressBar
	if progressBar {
		bar = progressbar.Default(int64(len(a.Results)))
	}

	for index, analysis := range a.Results {
		parsedText, err := a.AIClient.Parse(a.Context, analysis.Error, a.NoCache)
		if err != nil {
			// Check for exhaustion
			if strings.Contains(err.Error(), "status code: 429") {
				color.Red("Exhausted API quota. Please try again later")
				os.Exit(1)
			}
			color.Red("Error: %v", err)
			continue
		}
		analysis.Details = parsedText
		bar.Add(1)
		a.Results[index] = analysis
	}
	return nil
}
