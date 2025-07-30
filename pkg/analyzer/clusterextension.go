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
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ClusterExtensionAnalyzer struct{}

func (ClusterExtensionAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "ClusterExtension"

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	var clusterExtensionGVR = schema.GroupVersionResource{
		Group:    "olm.operatorframework.io",
		Version:  "v1",
		Resource: "clusterextensions",
	}
	if a.Client == nil {
		return nil, fmt.Errorf("client is nil in ClusterExtensionAnalyzer")
	}
	if a.Client.GetDynamicClient() == nil {
		return nil, fmt.Errorf("dynamic client is nil in ClusterExtensionAnalyzer")
	}

	list, err := a.Client.GetDynamicClient().Resource(clusterExtensionGVR).Namespace("").List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, item := range list.Items {
		var failures []common.Failure
		extension, err := ConvertToClusterExtension(&item)
		if err != nil {
			continue
		}
		fmt.Printf("ClusterExtension: %s | Source: %s\n", extension.Name, extension.Spec.Source.Catalog.PackageName)
		failures, err = ValidateClusterExtension(failures, extension)
		if err != nil {
			continue
		}

		if len(failures) > 0 {
			preAnalysis[extension.Name] = common.PreAnalysis{
				Extension:      *extension,
				FailureDetails: failures,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, extension.Name, "").Set(float64(len(failures)))
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, found := util.GetParent(a.Client, value.Node.ObjectMeta)
		if found {
			currentAnalysis.ParentObject = parent
		}
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, err
}

func ConvertToClusterExtension(u *unstructured.Unstructured) (*common.ClusterExtension, error) {
	var ce common.ClusterExtension
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &ce)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to ClusterExtension: %w", err)
	}
	return &ce, nil
}

func addExtensionConditionFailure(failures []common.Failure, extensionName string, extensionCondition metav1.Condition) []common.Failure {
	failures = append(failures, common.Failure{
		Text: fmt.Sprintf("OLMv1 ClusterExtension: %s has condition of type %s, reason %s: %s", extensionName, extensionCondition.Type, extensionCondition.Reason, extensionCondition.Message),
		Sensitive: []common.Sensitive{
			{
				Unmasked: extensionName,
				Masked:   util.MaskString(extensionName),
			},
		},
	})
	return failures
}

func addExtensionFailure(failures []common.Failure, extensionName string, err error) []common.Failure {
	failures = append(failures, common.Failure{
		Text: fmt.Sprintf("%s has error: %s", extensionName, err.Error()),
		Sensitive: []common.Sensitive{
			{
				Unmasked: extensionName,
				Masked:   util.MaskString(extensionName),
			},
		},
	})
	return failures
}

func ValidateClusterExtension(failures []common.Failure, extension *common.ClusterExtension) ([]common.Failure, error) {
	if extension.Spec.Source.Catalog != nil && extension.Spec.Source.Catalog.UpgradeConstraintPolicy != "CatalogProvided" && extension.Spec.Source.Catalog.UpgradeConstraintPolicy != "SelfCertified" {
		failures = addExtensionFailure(failures, extension.Name, fmt.Errorf("invalid or missing extension.Spec.Source.Catalog.UpgradeConstraintPolicy (expecting 'SelfCertified' or 'CatalogProvided')"))
	}

	if extension.Spec.Source.SourceType != "Catalog" {
		failures = addExtensionFailure(failures, extension.Name, fmt.Errorf("invalid or missing spec.source.sourceType (expecting 'Catalog')"))
	}

	for _, condition := range extension.Status.Conditions {
		if condition.Status != "True" && condition.Type == "Installed" {
			failures = addExtensionConditionFailure(failures, extension.Name, condition)
		}
		if condition.Type == "Progressing" && condition.Reason != "Succeeded" {
			failures = addExtensionConditionFailure(failures, extension.Name, condition)
		}
	}

	return failures, nil
}
