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

func TestPodAnalyzer(t *testing.T) {

	clientset := fake.NewSimpleClientset(
		&v1.Pod{
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
		},
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example2",
				Namespace: "default",
			},
			Status: v1.PodStatus{
				Phase: v1.PodRunning,
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name:  "example2",
						Ready: false,
					},
				},
				Conditions: []v1.PodCondition{
					{
						Type:    v1.ContainersReady,
						Reason:  "ContainersNotReady",
						Message: "containers with unready status: [example2]",
					},
				},
			},
		},
		// simulate event: 30s         Warning   Unhealthy              pod/my-nginx-7fb4dbcf47-4ch4w                         Readiness probe failed: bash: xxxx: command not found
		&v1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "default",
			},
			InvolvedObject: v1.ObjectReference{
				Kind:       "Pod",
				Name:       "example2",
				Namespace:  "default",
				UID:        "differentUid",
				APIVersion: "v1",
			},
			Reason:  "Unhealthy",
			Message: "readiness probe failed: the detail reason here ...",
			Source:  v1.EventSource{Component: "eventTest"},
			Count:   1,
			Type:    v1.EventTypeWarning,
		})

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	podAnalyzer := PodAnalyzer{}
	var analysisResults []common.Result
	analysisResults, err := podAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 2)
}

func TestPodAnalyzerNamespaceFiltering(t *testing.T) {

	clientset := fake.NewSimpleClientset(
		&v1.Pod{
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
		},
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "other-namespace",
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

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	podAnalyzer := PodAnalyzer{}
	var analysisResults []common.Result
	analysisResults, err := podAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 1)
}
