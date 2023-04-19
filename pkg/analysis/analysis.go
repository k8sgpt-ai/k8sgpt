package analysis

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
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
	Results   []common.Result
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
	Status   AnalysisStatus  `json:"status"`
	Problems int             `json:"problems"`
	Results  []common.Result `json:"results"`
}

func NewAnalysis(backend string, language string, filters []string, namespace string, noCache bool, explain bool) (*Analysis, error) {
	var configAI ai.AIConfiguration
	err := viper.UnmarshalKey("ai", &configAI)
	if err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}

	if len(configAI.Providers) == 0 && explain {
		color.Red("Error: AI provider not specified in configuration. Please run k8sgpt auth")
		os.Exit(1)
	}

	var aiProvider ai.AIProvider
	for _, provider := range configAI.Providers {
		if backend == provider.Name {
			aiProvider = provider
			break
		}
	}

	if aiProvider.Name == "" {
		color.Red("Error: AI provider %s not specified in configuration. Please run k8sgpt auth", backend)
		return nil, errors.New("AI provider not specified in configuration")
	}

	aiClient := ai.NewClient(aiProvider.Name)
	if err := aiClient.Configure(&aiProvider, language); err != nil {
		color.Red("Error: %v", err)
		return nil, err
	}

	ctx := context.Background()
	// Get kubernetes client from viper

	kubecontext := viper.GetString("kubecontext")
	kubeconfig := viper.GetString("kubeconfig")
	client, err := kubernetes.NewClient(kubecontext, kubeconfig)
	if err != nil {
		color.Red("Error initialising kubernetes client: %v", err)
		return nil, err
	}

	return &Analysis{
		Context:   ctx,
		Filters:   filters,
		Client:    client,
		AIClient:  aiClient,
		Namespace: namespace,
		NoCache:   noCache,
		Explain:   explain,
	}, nil
}

func (a *Analysis) RunAnalysis() error {
	activeFilters := viper.GetStringSlice("active_filters")

	analyzerMap := analyzer.GetAnalyzerMap()

	analyzerConfig := common.Analyzer{
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
			} else {
				return errors.New(fmt.Sprintf("\"%s\" filter does not exist. Please run k8sgpt filters list.", filter))
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
