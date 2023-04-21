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
	"strings"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/magiconair/properties/assert"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestIngressAnalyzer(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		})
	ingressAnalyzer := IngressAnalyzer{}

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	analysisResults, err := ingressAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 1)
}

func TestIngressAnalyzerWithMultipleIngresses(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		},
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example-2",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		},
	)
	ingressAnalyzer := IngressAnalyzer{}

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	analysisResults, err := ingressAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 2)
}

func TestIngressAnalyzerWithoutIngressClassAnnotation(t *testing.T) {

	clientset := fake.NewSimpleClientset(
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		})
	ingressAnalyzer := IngressAnalyzer{}

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	analysisResults, err := ingressAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}

	var errorFound bool
	for _, analysis := range analysisResults {
		for _, err := range analysis.Error {
			if strings.Contains(err.Text, "does not specify an Ingress class") {
				errorFound = true
				break
			}
		}
		if errorFound {
			break
		}
	}
	if !errorFound {
		t.Error("expected error 'does not specify an Ingress class' not found in analysis results")
	}
}

func TestIngressAnalyzerNamespaceFiltering(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		},
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "example",
				Namespace:   "other-namespace",
				Annotations: map[string]string{},
			},
		})
	ingressAnalyzer := IngressAnalyzer{}

	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: clientset,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	analysisResults, err := ingressAnalyzer.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 1)
}
