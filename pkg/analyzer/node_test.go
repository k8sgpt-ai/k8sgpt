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
	"context"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/magiconair/properties/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNodeAnalyzerNodeReady(t *testing.T) {
	clientset := fake.NewSimpleClientset(&v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Type:    v1.NodeReady,
					Status:  v1.ConditionTrue,
					Reason:  "KubeletReady",
					Message: "kubelet is posting ready status",
				},
			},
		},
	})

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context: context.Background(),
	}
	nodeAnalyzer := NodeAnalyzer{}
	var analysisResults []common.Result
	analysisResults, err := nodeAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 0)
}

func TestNodeAnalyzerNodeDiskPressure(t *testing.T) {
	clientset := fake.NewSimpleClientset(&v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Type:    v1.NodeDiskPressure,
					Status:  v1.ConditionTrue,
					Reason:  "KubeletHasDiskPressure",
					Message: "kubelet has disk pressure",
				},
			},
		},
	})

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context: context.Background(),
	}
	nodeAnalyzer := NodeAnalyzer{}
	var analysisResults []common.Result
	analysisResults, err := nodeAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 1)
}

// A cloud provider may set their own condition and/or a new status might be introduced
// In such cases a failure is assumed and the code shouldn't break, although it might be a false positive
func TestNodeAnalyzerNodeUnknownType(t *testing.T) {
	clientset := fake.NewSimpleClientset(&v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
		},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{
					Type:    "UnknownNodeConditionType",
					Status:  "CompletelyUnknown",
					Reason:  "KubeletHasTheUnknown",
					Message: "kubelet has the unknown",
				},
			},
		},
	})

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context: context.Background(),
	}
	nodeAnalyzer := NodeAnalyzer{}
	var analysisResults []common.Result
	analysisResults, err := nodeAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 1)
}
