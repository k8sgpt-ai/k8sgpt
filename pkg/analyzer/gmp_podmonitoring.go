/*
Copyright 2023 The K8sGPT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package analyzer

import (
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/gmp"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
)

type GMPPodMonitoringAnalyzer struct{}

func (g GMPPodMonitoringAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{"analyzer_name": "GMP.PodMonitoring"})
	// TODO(bwplotka): Pass the AnalyzerErrorsMetric metric and instrument.
	return gmp.AnalyzePodMonitorings(a)
}
