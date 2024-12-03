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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCRDSuccess(t *testing.T) {
	crd := apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "crdTest.stable.example.com",
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "stable.example.com",
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{{
				Name:    "v1alpha1",
				Served:  true,
				Storage: true,
			}},
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   "crdtests",
				Singular: "crdtest",
				Kind:     "CrdTest",
			},
			Scope: apiextensionsv1.ClusterScoped,
		},
	}

	scheme := scheme.Scheme

	err := apiextensionsv1.AddToScheme(scheme)
	if err != nil {
		t.Error(err)
	}

	objects := []runtime.Object{
		&crd,
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()

	analyzerInstance := CrdAnalyzer{}
	config := common.Analyzer{
		Client: &kubernetes.Client{
			CtrlClient: fakeClient,
			Config:     nil,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	analysisResults, err := analyzerInstance.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 0)
}

func TestCRDFailForConverstionWebhook(t *testing.T) {
	crd := apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "crdTest.stable.example.com",
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "stable.example.com",
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{{
				Name:    "v1alpha1",
				Served:  true,
				Storage: true,
			}},
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   "crdtests",
				Singular: "crdtest",
				Kind:     "CrdTest",
			},
			Scope: apiextensionsv1.ClusterScoped,
			Conversion: &apiextensionsv1.CustomResourceConversion{
				Strategy: "Webhook",
				Webhook: &apiextensionsv1.WebhookConversion{
					ClientConfig: &apiextensionsv1.WebhookClientConfig{
						Service: &apiextensionsv1.ServiceReference{
							Name:      "example-conversion-webhook-server",
							Namespace: "default",
						},
					},
				},
			},
		},
	}

	scheme := scheme.Scheme

	err := apiextensionsv1.AddToScheme(scheme)
	if err != nil {
		t.Error(err)
	}

	objects := []runtime.Object{
		&crd,
	}

	fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()

	analyzerInstance := CrdAnalyzer{}
	clientset := fake.NewSimpleClientset()
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client:     clientset,
			CtrlClient: fakeClient,
			Config:     nil,
		},
		Context:   context.Background(),
		Namespace: "default",
	}
	analysisResults, err := analyzerInstance.Analyze(config)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(analysisResults), 1)
}
