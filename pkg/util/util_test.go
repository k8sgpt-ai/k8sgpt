/*
Copyright 2024 The K8sGPT Authors.
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

package util

import (
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/stretchr/testify/require"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestSliceContainsString(t *testing.T) {
	tests := []struct {
		slice    []string
		s        string
		expected bool
	}{
		{
			expected: false,
		},
		{
			slice:    []string{"temp", "value"},
			s:        "value",
			expected: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.s, func(t *testing.T) {
			require.Equal(t, tt.expected, SliceContainsString(tt.slice, tt.s))
		})
	}
}

func TestGetParent(t *testing.T) {
	ownerName := "test-name"
	namespace := "test"
	clientset := fake.NewSimpleClientset(
		&appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ownerName,
				Namespace: namespace,
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ownerName,
				Namespace: namespace,
			},
		},
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ownerName,
				Namespace: namespace,
			},
		},
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ownerName,
				Namespace: namespace,
			},
		},
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ownerName,
				Namespace: namespace,
			},
		},
		&admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: ownerName,
			},
		},
		&admissionregistrationv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: ownerName,
			},
		},
	)
	kubeClient := kubernetes.Client{
		Client: clientset,
	}

	tests := []struct {
		name           string
		kind           string
		expectedOutput string
	}{
		{
			kind:           "Unknown",
			expectedOutput: ownerName,
		},
		{
			kind: "ReplicaSet",
		},
		{
			kind:           "ReplicaSet",
			name:           ownerName,
			expectedOutput: "ReplicaSet/test-name",
		},
		{
			kind: "Deployment",
		},
		{
			kind:           "Deployment",
			name:           ownerName,
			expectedOutput: "Deployment/test-name",
		},
		{
			kind: "StatefulSet",
		},
		{
			kind:           "StatefulSet",
			name:           ownerName,
			expectedOutput: "StatefulSet/test-name",
		},
		{
			kind: "DaemonSet",
		},
		{
			kind:           "DaemonSet",
			name:           ownerName,
			expectedOutput: "DaemonSet/test-name",
		},
		{
			kind: "Ingress",
		},
		{
			kind:           "Ingress",
			name:           ownerName,
			expectedOutput: "Ingress/test-name",
		},
		{
			kind: "MutatingWebhookConfiguration",
		},
		{
			kind:           "MutatingWebhookConfiguration",
			name:           ownerName,
			expectedOutput: "MutatingWebhook/test-name",
		},
		{
			kind: "ValidatingWebhookConfiguration",
		},
		{
			kind:           "ValidatingWebhookConfiguration",
			name:           ownerName,
			expectedOutput: "ValidatingWebhook/test-name",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.kind, func(t *testing.T) {
			meta := metav1.ObjectMeta{
				Namespace: namespace,
				Name:      ownerName,
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: tt.kind,
						Name: tt.name,
					},
				},
			}
			output, ok := GetParent(&kubeClient, meta)
			require.Equal(t, tt.expectedOutput, output)
			require.Equal(t, false, ok)
		})
	}
}

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name               string
		slice              []string
		expectedDuplicates []string
	}{
		{
			name:               "all empty",
			expectedDuplicates: []string{},
		},
		{
			name:               "all unique",
			slice:              []string{"temp", "value"},
			expectedDuplicates: []string{},
		},
		{
			name:               "slice not unique",
			slice:              []string{"temp", "mango", "mango"},
			expectedDuplicates: []string{"mango"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, duplicates := RemoveDuplicates(tt.slice)
			require.Equal(t, tt.expectedDuplicates, duplicates)
		})
	}
}

func TestSliceDiff(t *testing.T) {
	tests := []struct {
		name         string
		source       []string
		dest         []string
		expectedDiff []string
	}{
		{
			name: "all empty",
		},
		{
			name:         "non empty",
			source:       []string{"temp"},
			dest:         []string{"random"},
			expectedDiff: []string{"temp"},
		},
		{
			name:   "no diff",
			source: []string{"temp", "random"},
			dest:   []string{"random", "temp"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expectedDiff, SliceDiff(tt.source, tt.dest))
		})
	}
}

func TestReplaceIfMatch(t *testing.T) {
	tests := []struct {
		text           string
		pattern        string
		replacement    string
		expectedOutput string
	}{
		{
			text:           "",
			expectedOutput: "",
		},
		{
			text:           "some value",
			pattern:        "new",
			replacement:    "latest",
			expectedOutput: "some value",
		},
		{
			text:           "new value",
			pattern:        "value",
			replacement:    "day",
			expectedOutput: "new day",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.text, func(t *testing.T) {
			require.Equal(t, tt.expectedOutput, ReplaceIfMatch(tt.text, tt.pattern, tt.replacement))
		})
	}
}

func TestGetCacheKey(t *testing.T) {
	tests := []struct {
		provider       string
		language       string
		sEnc           string
		expectedOutput string
	}{
		{
			expectedOutput: "d8156bae0c4243d3742fc4e9774d8aceabe0410249d720c855f98afc88ff846c",
		},
		{
			provider:       "provider",
			language:       "english",
			sEnc:           "encoding",
			expectedOutput: "39415cc324b1553b93e80e46049e4e4dbb752dc7d0424b2c6ac96d745c6392aa",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.language, func(t *testing.T) {
			require.Equal(t, tt.expectedOutput, GetCacheKey(tt.provider, tt.language, tt.sEnc))
		})
	}
}

func TestGetPodListByLabels(t *testing.T) {
	namespace1 := "test1"
	namespace2 := "test2"
	clientset := fake.NewSimpleClientset(
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "Pod1",
				Namespace: namespace1,
				Labels: map[string]string{
					"Name":      "Pod1",
					"Namespace": namespace1,
				},
			},
		},
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "Pod2",
				Namespace: namespace2,
				Labels: map[string]string{
					"Name":      "Pod2",
					"Namespace": namespace2,
				},
			},
		},
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "Pod3",
				Namespace: namespace1,
				Labels: map[string]string{
					"Name":      "Pod3",
					"Namespace": namespace1,
				},
			},
		},
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "Pod4",
				Namespace: namespace2,
				Labels: map[string]string{
					"Name":      "Pod4",
					"Namespace": namespace2,
				},
			},
		},
	)

	tests := []struct {
		name        string
		namespace   string
		labels      map[string]string
		expectedLen int
		expectedErr string
	}{
		{
			name:      "Name is Pod1",
			namespace: namespace1,
			labels: map[string]string{
				"Name": "Pod1",
			},
			expectedLen: 1,
		},
		{
			name:      "Name is Pod2 in namespace1",
			namespace: namespace1,
			labels: map[string]string{
				"Name": "Pod2",
			},
			expectedLen: 0,
		},
		{
			name:      "Name is Pod2 in namespace 2",
			namespace: namespace2,
			labels: map[string]string{
				"Name": "Pod2",
			},
			expectedLen: 1,
		},
		{
			name:      "All pods with namespace2 label in namespace1",
			namespace: namespace1,
			labels: map[string]string{
				"Namespace": namespace2,
			},
			expectedLen: 0,
		},
		{
			name:      "All pods with namespace2 label in namespace2",
			namespace: namespace2,
			labels: map[string]string{
				"Namespace": namespace2,
			},
			expectedLen: 2,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			pl, err := GetPodListByLabels(clientset, tt.namespace, tt.labels)
			if tt.expectedErr == "" {
				require.NoError(t, err)
				require.Equal(t, tt.expectedLen, len(pl.Items))
			} else {
				require.ErrorContains(t, err, tt.expectedErr)
				require.Nil(t, pl)
			}
		})
	}
}
func TestFileExists(t *testing.T) {
	tests := []struct {
		filePath  string
		isPresent bool
		err       string
	}{
		{
			filePath:  "",
			isPresent: false,
		},
		{
			filePath:  "./util.go",
			isPresent: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.filePath, func(t *testing.T) {
			isPresent, err := FileExists(tt.filePath)
			if tt.err == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.err)
			}
			require.Equal(t, tt.isPresent, isPresent)
		})
	}
}

func TestEnsureDirExists(t *testing.T) {
	tests := []struct {
		dir string
		err string
	}{
		{
			dir: "",
			err: "mkdir : no such file or directory",
		},
		{
			dir: "./",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.dir, func(t *testing.T) {
			err := EnsureDirExists(tt.dir)
			if tt.err == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.err)
			}
		})
	}
}

func TestMapToString(t *testing.T) {
	tests := []struct {
		name   string
		m      map[string]string
		output string
	}{
		{
			name: "empty map",
			m:    map[string]string{},
		},
		{
			name: "non-empty map",
			m: map[string]string{
				"key": "value",
			},
			output: "key=value",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.output, MapToString(tt.m))
		})
	}
}

func TestLabelsIncludeAny(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]string
		p    map[string]string
		ok   bool
	}{
		{
			name: "empty map",
			m:    map[string]string{},
			p:    map[string]string{},
			ok:   false,
		},
		{
			name: "non-empty map",
			m: map[string]string{
				"key": "value",
			},
			p: map[string]string{
				"key": "value",
			},
			ok: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.ok, LabelsIncludeAny(tt.p, tt.m))
		})
	}
}
