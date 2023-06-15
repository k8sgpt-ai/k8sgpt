package alex

import "github.com/k8sgpt-ai/k8sgpt/pkg/common"

type AlexAnalyzer struct {
}

func NewAlexAnalyzer() *AlexAnalyzer {
	return &AlexAnalyzer{}
}

func (*AlexAnalyzer) Analyze(analysis common.Analyzer) ([]common.Result, error) {

	return nil, nil
}
