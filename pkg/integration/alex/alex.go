package alex

import (
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
)

// implement the IIterator interface
type Alex struct {
}

func NewAlex() *Alex {
	return &Alex{}
}

func (a *Alex) Deploy(namespace string) error {
	return nil
}

func (a *Alex) UnDeploy(namespace string) error {
	return nil
}

func (a *Alex) AddAnalyzer(analyzers *map[string]common.IAnalyzer) {

	(*analyzers)["Alex"] = NewAlexAnalyzer()
}

func (a *Alex) RemoveAnalyzer() error {

	return nil
}

func (a *Alex) GetAnalyzerName() string {

	return "alex"
}

func (a *Alex) IsActivate() bool {

	return true
}
