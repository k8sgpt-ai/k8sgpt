/*
Copyright 2026 The K8sGPT Authors.
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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NodeClaimAnalyzer reports Karpenter v1 NodeClaims whose Launched,
// Registered or Initialized condition is False, which usually means the
// node a pending pod is waiting for will never become ready.
type NodeClaimAnalyzer struct{}

var nodeClaimGVR = schema.GroupVersionResource{
	Group:    "karpenter.sh",
	Version:  "v1",
	Resource: "nodeclaims",
}

// nodeClaimLifecycleConditions are the NodeClaim conditions that must become
// True for a node to join the cluster and accept pods.
var nodeClaimLifecycleConditions = map[string]struct{}{
	"Launched":    {},
	"Registered":  {},
	"Initialized": {},
}

func (NodeClaimAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	const kind = "NodeClaim"

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	if a.Client == nil || a.Client.GetDynamicClient() == nil {
		return nil, fmt.Errorf("dynamic client is nil in %s analyzer", kind)
	}

	// NodeClaims are cluster-scoped, so a.Namespace is intentionally not applied.
	list, err := a.Client.GetDynamicClient().
		Resource(nodeClaimGVR).
		List(a.Context, a.ListOptions())
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("karpenter.sh/v1 NodeClaim CRD not found; is Karpenter installed?: %w", err)
		}
		return nil, err
	}

	var results []common.Result
	for _, item := range list.Items {
		name := item.GetName()

		conditions, found, err := unstructured.NestedSlice(item.Object, "status", "conditions")
		if err != nil || !found {
			// No status yet (just created) or a malformed object: nothing to report.
			continue
		}

		var failures []common.Failure
		for _, elem := range conditions {
			condition, ok := elem.(map[string]interface{})
			if !ok {
				continue
			}

			// type and status are required; reason and message are optional.
			conditionType, ok := condition["type"].(string)
			if !ok {
				continue
			}
			status, ok := condition["status"].(string)
			if !ok {
				continue
			}
			if _, tracked := nodeClaimLifecycleConditions[conditionType]; !tracked || status != "False" {
				continue
			}
			reason, _ := condition["reason"].(string)
			message, _ := condition["message"].(string)

			text := fmt.Sprintf("NodeClaim %s condition %s=False", name, conditionType)
			if reason != "" {
				text += ": " + reason
			}
			if message != "" {
				text += ": " + message
			}

			failures = append(failures, common.Failure{
				Text: text,
				Sensitive: []common.Sensitive{
					{
						Unmasked: name,
						Masked:   util.MaskString(name),
					},
				},
			})
		}

		if len(failures) == 0 {
			continue
		}

		results = append(results, common.Result{
			Kind:  kind,
			Name:  name,
			Error: failures,
		})
		AnalyzerErrorsMetric.WithLabelValues(kind, name, "").Set(float64(len(failures)))
	}

	return results, nil
}
