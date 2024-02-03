package analyzer_test

import (
	"context"
	"testing"
	"time"

	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/magiconair/properties/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestFetchLatestEvent(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	event1 := &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-event-1",
			Namespace: "default",
		},
		LastTimestamp: metav1.Time{Time: time.Now()},
	}
	event2 := &v1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-event-2",
			Namespace: "default",
		},
		LastTimestamp: metav1.Time{Time: time.Now().Add(-time.Hour)},
	}

	// Create the namespace and add the events to the same
	_, err := fakeClient.CoreV1().Namespaces().Create(context.TODO(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "default"},
	}, metav1.CreateOptions{})
	assert.Equal(t, err, nil)
	_, err = fakeClient.CoreV1().Events("default").Create(context.TODO(), event1, metav1.CreateOptions{})
	assert.Equal(t, err, nil)
	_, err = fakeClient.CoreV1().Events("default").Create(context.TODO(), event2, metav1.CreateOptions{})
	assert.Equal(t, err, nil)

	tests := []struct {
		name       string
		namespace  string
		nameToFind string
		expected   *v1.Event
		shouldFail bool
	}{
		// Should fetch the latest event if both the namespace and event exist
		{
			name:       "Valid case",
			namespace:  "default",
			nameToFind: "test-event",
			expected:   event1,
			shouldFail: false,
		},
		// Should fail if the event does not exist
		{
			name:       "Nonexistent event",
			namespace:  "default",
			nameToFind: "nonexistent-event",
			expected:   nil,
			shouldFail: true,
		},
		// Should return nl if the namespace does not exist
		{
			name:       "Nonexistent namespace",
			namespace:  "nonexistent-namespace",
			nameToFind: "test-event",
			expected:   nil,
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := kubernetes.Client{
				Client: fakeClient,
			}
			assert.Equal(t, err, nil)

			analyzer, err := analyzer.FetchLatestEvent(context.TODO(), &client, tt.namespace, tt.nameToFind)
			assert.Equal(t, err, nil)

			if tt.shouldFail && analyzer == nil {
				t.Error("Expected an error, but got nil")
			}
			if !tt.shouldFail && analyzer != nil && analyzer.Name != tt.expected.Name {
				t.Errorf("Expected event name %s, got %s", tt.expected.Name, analyzer.Name)
			}
		})
	}
}
