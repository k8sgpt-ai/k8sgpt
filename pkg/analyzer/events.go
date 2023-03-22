package analyzer

import (
	"context"

	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func FetchLatestPodEvent(ctx context.Context, kubernetesClient *kubernetes.Client, pod *v1.Pod) (*v1.Event, error) {

	// get the list of events
	events, err := kubernetesClient.GetClient().CoreV1().Events(pod.Namespace).List(ctx,
		metav1.ListOptions{
			FieldSelector: "involvedObject.name=" + pod.Name,
		})

	if err != nil {
		return nil, err
	}
	// find most recent event
	var latestEvent *v1.Event
	for _, event := range events.Items {
		if latestEvent == nil {
			latestEvent = &event
		}
		if event.LastTimestamp.After(latestEvent.LastTimestamp.Time) {
			latestEvent = &event
		}
	}
	return latestEvent, nil
}
