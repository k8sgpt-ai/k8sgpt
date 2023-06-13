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

package netobserv

import (
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	"github.com/fatih/color"
	flowslatest "github.com/netobserv/network-observability-operator/api/v1beta1"
	"k8s.io/client-go/rest"
)

type NetObservAnalyzer struct {
}

func (NetObservAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	result := &flowslatest.FlowCollectorList{}

	config := a.Client.GetConfig()
	// Add group version to sceheme
	config.ContentConfig.GroupVersion = &flowslatest.GroupVersion
	config.UserAgent = rest.DefaultKubernetesUserAgent()
	config.APIPath = "/apis"

	restClient, err := rest.UnversionedRESTClientFor(config)
	if err != nil {
		color.Red("failed to get config err: %v", err)
		return nil, err
	}
	err = restClient.Get().Resource("flowcollectors.flows.netobserv.io").Do(a.Context).Into(result)
	if err != nil {
		color.Red("failed to get flowcollector resource err: %v", err)
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	color.Green("Object %+v\n", result)
	for _, report := range result.Items {
		color.Green("Object %+v\n", report)
		var failures []common.Failure
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", report.Labels["netobserv-operator.resource.namespace"],
				report.Labels["netobserv-operator.resource.name"])] = common.PreAnalysis{
				NetobservReport: report,
				FailureDetails:   failures,
			}
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  "NetObservrReport",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.TrivyVulnerabilityReport.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}
