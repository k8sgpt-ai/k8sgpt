package analyzer_test

import (
	"context"
	"errors"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func FetchLatestEvent(ctx context.Context, client kubernetes.Interface, namespace, eventName string) (*v1.Event, error) {
	// List events in the specified namespace
	events, err := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var latestEvent *v1.Event
	for _, event := range events.Items {
		// Check if the event name matches the requested name (partial match)
		if eventName == "" || event.Name == eventName {
			if latestEvent == nil || event.LastTimestamp.Time.After(latestEvent.LastTimestamp.Time) {
				latestEvent = &event
			}
		}
	}

	// If no matching event is found, return an error
	if latestEvent == nil {
		return nil, errors.New("no matching events found")
	}
	return latestEvent, nil
}
func TestFetchLatestEvent(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	// Simulating events with different timestamps
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
		LastTimestamp: metav1.Time{Time: time.Now().Add(-time.Hour)}, // event1 should be fetched as it's newer
	}

	// ✅ Explicitly ensure namespace exists
	_, err := fakeClient.CoreV1().Namespaces().Create(context.TODO(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "default"},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create namespace: %v", err)
	}

	// ✅ Ensure events are properly created and stored in the fake client
	_, err = fakeClient.CoreV1().Events("default").Create(context.TODO(), event1, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create event1: %v", err)
	}

	_, err = fakeClient.CoreV1().Events("default").Create(context.TODO(), event2, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create event2: %v", err)
	}

	// 🔍 Debug: Check if events exist before running FetchLatestEvent
	storedEvents, _ := fakeClient.CoreV1().Events("default").List(context.TODO(), metav1.ListOptions{})
	if len(storedEvents.Items) == 0 {
		t.Fatal("No events were found in the fake client. Ensure event creation is working correctly.")
	}

	// Test cases
	tests := []struct {
		name       string
		namespace  string
		nameToFind string
		expected   *v1.Event
		shouldFail bool
	}{
		{
			name:       "Valid case - fetch the latest event",
			namespace:  "default",
			nameToFind: "test-event-1", // Match exact event name
			expected:   event1,         // event1 has the latest timestamp
			shouldFail: false,
		},
		{
			name:       "Nonexistent event",
			namespace:  "default",
			nameToFind: "nonexistent-event", // Should not exist
			expected:   nil,
			shouldFail: true,
		},
		{
			name:       "Nonexistent namespace",
			namespace:  "nonexistent-namespace", // Namespace doesn't exist
			nameToFind: "test-event",
			expected:   nil,
			shouldFail: true,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function to fetch the latest event
			event, err := FetchLatestEvent(context.TODO(), fakeClient, tt.namespace, tt.nameToFind)

			// Handle the expected outcomes based on the test case
			if tt.shouldFail {
				if err == nil {
					t.Error("Expected an error, but got nil")
				}
				if event != nil {
					t.Errorf("Expected nil event, but got event: %s", event.Name)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got %v", err)
				}
				if event != nil && event.Name != tt.expected.Name {
					t.Errorf("Expected event name %s, got %s", tt.expected.Name, event.Name)
				}
			}
		})
	}
}
