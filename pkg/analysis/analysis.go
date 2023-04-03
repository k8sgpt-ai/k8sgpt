package analysis

import (
	"context"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
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
