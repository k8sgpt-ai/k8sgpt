package analyzer

import (
	"context"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/magiconair/properties/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPodAnalzyer(t *testing.T) {

	clientset := fake.NewSimpleClientset(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "example",
			Namespace:   "default",
			Annotations: map[string]string{},
		},
		Status: v1.PodStatus{
			Phase: v1.PodPending,
			Conditions: []v1.PodCondition{
				{
					Type:    v1.PodScheduled,
					Reason:  "Unschedulable",
					Message: "0/1 nodes are available: 1 node(s) had taint {node-role.kubernetes.io/master: }, that the pod didn't tolerate.",
				},
			},
		},
	})

	podAnalyzer := PodAnalyzer{}
	var analysisResults []Analysis
	podAnalyzer.RunAnalysis(context.Background(),
		&AnalysisConfiguration{
			Namespace: "default",
		},
		&kubernetes.Client{
			Client: clientset,
		}, nil, &analysisResults)

	assert.Equal(t, len(analysisResults), 1)
}
