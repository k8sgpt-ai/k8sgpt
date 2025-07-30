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
	"regexp"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ClusterCatalogAnalyzer struct{}

func (ClusterCatalogAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	kind := "ClusterCatalog"

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	var clusterCatalogGVR = schema.GroupVersionResource{
		Group:    "olm.operatorframework.io",
		Version:  "v1",
		Resource: "clustercatalogs",
	}
	if a.Client == nil {
		return nil, fmt.Errorf("client is nil in ClusterCatalogAnalyzer")
	}
	if a.Client.GetDynamicClient() == nil {
		return nil, fmt.Errorf("dynamic client is nil in ClusterCatalogAnalyzer")
	}

	list, err := a.Client.GetDynamicClient().Resource(clusterCatalogGVR).Namespace("").List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, item := range list.Items {
		var failures []common.Failure
		catalog, err := ConvertToClusterCatalog(&item)
		if err != nil {
			continue
		}
		fmt.Printf("ClusterCatalog: %s | Source: %s\n", catalog.Name, catalog.Spec.Source.Image.Ref)
		failures, err = ValidateClusterCatalog(failures, catalog)
		if err != nil {
			continue
		}

		if len(failures) > 0 {
			preAnalysis[catalog.Name] = common.PreAnalysis{
				Catalog:        *catalog,
				FailureDetails: failures,
			}
			AnalyzerErrorsMetric.WithLabelValues(kind, catalog.Name, "").Set(float64(len(failures)))
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

func ConvertToClusterCatalog(u *unstructured.Unstructured) (*common.ClusterCatalog, error) {
	var cc common.ClusterCatalog
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &cc)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to ClusterCatalog: %w", err)
	}
	return &cc, nil
}

func addCatalogConditionFailure(failures []common.Failure, catalogName string, catalogCondition metav1.Condition) []common.Failure {
	failures = append(failures, common.Failure{
		Text: fmt.Sprintf("OLMv1 ClusterCatalog: %s has condition of type %s, reason %s: %s", catalogName, catalogCondition.Type, catalogCondition.Reason, catalogCondition.Message),
		Sensitive: []common.Sensitive{
			{
				Unmasked: catalogName,
				Masked:   util.MaskString(catalogName),
			},
		},
	})
	return failures
}

func addCatalogFailure(failures []common.Failure, catalogName string, err error) []common.Failure {
	failures = append(failures, common.Failure{
		Text: fmt.Sprintf("%s has error: %s", catalogName, err.Error()),
		Sensitive: []common.Sensitive{
			{
				Unmasked: catalogName,
				Masked:   util.MaskString(catalogName),
			},
		},
	})
	return failures
}

func ValidateClusterCatalog(failures []common.Failure, catalog *common.ClusterCatalog) ([]common.Failure, error) {
	if !isValidImageRef(catalog.Spec.Source.Image.Ref) {
		failures = addCatalogFailure(failures, catalog.Name, fmt.Errorf("invalid image ref format in spec.source.image.ref: %s", catalog.Spec.Source.Image.Ref))
	}

	// Check status.resolvedSource.image.ref ends with @sha256:...
	if catalog.Status.ResolvedSource != nil {
		if catalog.Status.ResolvedSource.Image.Ref == "" {
			failures = addCatalogFailure(failures, catalog.Name, fmt.Errorf("missing status.resolvedSource.image.ref"))
		}
		if !regexp.MustCompile(`@sha256:[a-f0-9]{64}$`).MatchString(catalog.Status.ResolvedSource.Image.Ref) {
			failures = addCatalogFailure(failures, catalog.Name, fmt.Errorf("status.resolvedSource.image.ref must end with @sha256:<digest>"))
		}
	}

	for _, condition := range catalog.Status.Conditions {
		if condition.Status != "True" && condition.Type == "Serving" {
			failures = addCatalogConditionFailure(failures, catalog.Name, condition)
		}
		if condition.Type == "Progressing" && condition.Reason != "Succeeded" {
			failures = addCatalogConditionFailure(failures, catalog.Name, condition)
		}
	}

	return failures, nil
}

// isValidImageRef does a simple regex check to validate image refs
func isValidImageRef(ref string) bool {
	pattern := `^([a-zA-Z0-9\-\.]+(?::[0-9]+)?/)?([a-z0-9]+(?:[._\-\/][a-z0-9]+)*)(:[\w][\w.-]{0,127})?(?:@sha256:[a-f0-9]{64})?$`
	return regexp.MustCompile(pattern).MatchString(ref)
}
