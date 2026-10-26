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
	"sort"
	"strconv"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// NodePoolAnalyzer reports Karpenter v1 NodePools that cannot provision new
// nodes: pools whose Ready condition is False, and pools whose current usage
// has reached one of their spec.limits. Both commonly leave pods Pending with
// no error on the Pod or Node objects.
type NodePoolAnalyzer struct{}

var nodePoolGVR = schema.GroupVersionResource{
	Group:    "karpenter.sh",
	Version:  "v1",
	Resource: "nodepools",
}

func (NodePoolAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	const kind = "NodePool"

	AnalyzerErrorsMetric.DeletePartialMatch(map[string]string{
		"analyzer_name": kind,
	})

	if a.Client == nil || a.Client.GetDynamicClient() == nil {
		return nil, fmt.Errorf("dynamic client is nil in %s analyzer", kind)
	}

	// NodePools are cluster-scoped, so a.Namespace is intentionally not applied.
	list, err := a.Client.GetDynamicClient().
		Resource(nodePoolGVR).
		List(a.Context, a.ListOptions())
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("karpenter.sh/v1 NodePool: CRD not installed: %w", err)
		}
		return nil, err
	}

	var results []common.Result
	for _, item := range list.Items {
		name := item.GetName()
		sensitive := []common.Sensitive{
			{
				Unmasked: name,
				Masked:   util.MaskString(name),
			},
		}

		var failures []common.Failure
		for _, text := range nodePoolReadyFailures(name, item.Object) {
			failures = append(failures, common.Failure{Text: text, Sensitive: sensitive})
		}
		for _, text := range nodePoolLimitFailures(name, item.Object) {
			failures = append(failures, common.Failure{Text: text, Sensitive: sensitive})
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

// nodePoolReadyFailures reports a Ready condition that is False. Unknown is
// not reported, because it is the normal state while a NodePool is reconciled.
func nodePoolReadyFailures(name string, obj map[string]interface{}) []string {
	conditions, found, err := unstructured.NestedSlice(obj, "status", "conditions")
	if err != nil || !found {
		return nil
	}

	var failures []string
	for _, elem := range conditions {
		condition, ok := elem.(map[string]interface{})
		if !ok {
			continue
		}
		conditionType, _ := condition["type"].(string)
		status, _ := condition["status"].(string)
		if conditionType != "Ready" || status != "False" {
			continue
		}
		reason, _ := condition["reason"].(string)
		message, _ := condition["message"].(string)

		text := fmt.Sprintf("NodePool %s condition Ready=False", name)
		if reason != "" {
			text += ": " + reason
		}
		if message != "" {
			text += ": " + message
		}
		failures = append(failures, text)
	}
	return failures
}

// nodePoolLimitFailures reports every resource in spec.limits whose usage in
// status.resources has reached the limit. Once a limit is reached Karpenter
// stops launching nodes for the pool. Limits of zero are skipped, because they
// are used deliberately to stop a pool from provisioning.
func nodePoolLimitFailures(name string, obj map[string]interface{}) []string {
	limits, found, err := unstructured.NestedMap(obj, "spec", "limits")
	if err != nil || !found || len(limits) == 0 {
		return nil
	}
	usage, found, err := unstructured.NestedMap(obj, "status", "resources")
	if err != nil || !found {
		return nil
	}

	// Sort for deterministic output.
	resourceNames := make([]string, 0, len(limits))
	for resourceName := range limits {
		resourceNames = append(resourceNames, resourceName)
	}
	sort.Strings(resourceNames)

	var failures []string
	for _, resourceName := range resourceNames {
		limit, ok := toQuantity(limits[resourceName])
		if !ok || limit.IsZero() {
			continue
		}
		used, ok := toQuantity(usage[resourceName])
		if !ok {
			continue
		}
		if used.Cmp(limit) < 0 {
			continue
		}
		failures = append(failures, fmt.Sprintf(
			"NodePool %s has reached its %s limit (%s used of %s); Karpenter will not launch new nodes for this NodePool",
			name, resourceName, used.String(), limit.String(),
		))
	}
	return failures
}

// toQuantity converts a value read from an unstructured object into a
// resource.Quantity. Quantities are normally strings ("1000", "4000Gi"), but
// plain JSON numbers are accepted too.
func toQuantity(v interface{}) (resource.Quantity, bool) {
	switch val := v.(type) {
	case string:
		q, err := resource.ParseQuantity(val)
		if err != nil {
			return resource.Quantity{}, false
		}
		return q, true
	case int64:
		return *resource.NewQuantity(val, resource.DecimalSI), true
	case float64:
		q, err := resource.ParseQuantity(strconv.FormatFloat(val, 'f', -1, 64))
		if err != nil {
			return resource.Quantity{}, false
		}
		return q, true
	default:
		return resource.Quantity{}, false
	}
}
