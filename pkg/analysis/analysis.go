package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/spf13/viper"
)

type Analysis struct {
	Context         context.Context
	Namespace       string
	NoCache         bool
	Explain         bool
	AIClient        ai.IAI
	Filters         []string
	Client          *kubernetes.Client
	analysisResults []common.Result
}

func NewAnalysis(namespace string, noCache bool, explain bool, filters []string, aiProvider string) *Analysis {
	var aiClient ai.IAI
	var err error

	ctx := context.Background()
	client := viper.Get("kubernetesClient").(*kubernetes.Client)

	if explain {
		aiClient, err = ai.NewAIClient(aiProvider)
		if err != nil {
			fmt.Println("Error creating AI client: ", err)
		}
	}

	return &Analysis{
		Context:   ctx,
		Namespace: namespace,
		NoCache:   noCache,
		Explain:   explain,
		Filters:   filters,
		Client:    client,
		AIClient:  aiClient,
	}
}

func (a *Analysis) RunAnalysis() error {
	activeFilters := viper.GetStringSlice("active_filters")
	analyzerList := analyzer.GetAnalyzerList()

	// if there are no filters selected and no active_filters then run all of them
	if len(a.Filters) == 0 && len(activeFilters) == 0 {
		for _, al := range analyzerList {
			thisanalysis, _ := analyzer.NewAnalyzer(al, a.Client, a.Context, a.Namespace, a.AIClient, a.Explain)
			err := thisanalysis.Analyze()
			if err != nil {
				fmt.Println("Error running analysis: ", err)
			}
			a.analysisResults = append(a.analysisResults, thisanalysis.GetResult()...)
		}
		return nil
	}

	// if the filters flag is specified
	if len(a.Filters) != 0 {
		for _, filter := range a.Filters {
			for _, ali := range analyzerList {
				if filter == ali {
					thisanalysis, _ := analyzer.NewAnalyzer(ali, a.Client, a.Context, a.Namespace, a.AIClient, a.Explain)
					err := thisanalysis.Analyze()
					if err != nil {
						fmt.Println("Error running analysis: ", err)
					}
					a.analysisResults = append(a.analysisResults, thisanalysis.GetResult()...)
				}
			}
		}
		return nil
	}
	return nil
}

func (a *Analysis) PrintAnalysisResult() {
	for _, result := range a.analysisResults {
		fmt.Println(result)
	}
}

func (a *Analysis) PrintJsonResult() {
	output, err := json.MarshalIndent(a.analysisResults, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling json: ", err)
	}
	fmt.Println(string(output))
}
