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
	"context"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/stretchr/testify/require"
	resourcev1beta1 "k8s.io/api/resource/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestResourceClaimAnalyzerReportsMissingDeviceClass(t *testing.T) {
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: fake.NewSimpleClientset(
				&resourcev1beta1.DeviceClass{
					ObjectMeta: metav1.ObjectMeta{Name: "gpu"},
				},
				&resourcev1beta1.ResourceClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "workload-devices",
						Namespace: "default",
					},
					Spec: resourcev1beta1.ResourceClaimSpec{
						Devices: resourcev1beta1.DeviceClaim{
							Requests: []resourcev1beta1.DeviceRequest{
								{Name: "accelerator", DeviceClassName: "gpu"},
								{Name: "network", DeviceClassName: "missing-class"},
								{Name: "network-copy", DeviceClassName: "missing-class"},
							},
						},
					},
				},
				&resourcev1beta1.ResourceClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "no-device-requests",
						Namespace: "default",
					},
				},
			),
		},
		Context:   context.Background(),
		Namespace: "default",
	}

	results, err := (ResourceClaimAnalyzer{}).Analyze(config)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "ResourceClaim", results[0].Kind)
	require.Equal(t, "default/workload-devices", results[0].Name)
	require.Len(t, results[0].Error, 1)
	require.Equal(t,
		"ResourceClaim workload-devices references DeviceClass missing-class which does not exist.",
		results[0].Error[0].Text,
	)
}

func TestResourceClaimAnalyzerAppliesClaimSelectorOnly(t *testing.T) {
	config := common.Analyzer{
		Client: &kubernetes.Client{
			Client: fake.NewSimpleClientset(
				&resourcev1beta1.DeviceClass{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "gpu",
						Labels: map[string]string{"tier": "platform"},
					},
				},
				&resourcev1beta1.ResourceClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "selected-claim",
						Namespace: "default",
						Labels:    map[string]string{"app": "selected"},
					},
					Spec: resourcev1beta1.ResourceClaimSpec{
						Devices: resourcev1beta1.DeviceClaim{
							Requests: []resourcev1beta1.DeviceRequest{
								{Name: "accelerator", DeviceClassName: "gpu"},
							},
						},
					},
				},
				&resourcev1beta1.ResourceClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unselected-claim",
						Namespace: "default",
						Labels:    map[string]string{"app": "other"},
					},
					Spec: resourcev1beta1.ResourceClaimSpec{
						Devices: resourcev1beta1.DeviceClaim{
							Requests: []resourcev1beta1.DeviceRequest{
								{Name: "accelerator", DeviceClassName: "missing-class"},
							},
						},
					},
				},
			),
		},
		Context:       context.Background(),
		Namespace:     "default",
		LabelSelector: "app=selected",
	}

	results, err := (ResourceClaimAnalyzer{}).Analyze(config)
	require.NoError(t, err)
	require.Empty(t, results)
}

func TestResourceClaimAnalyzerAppliesResourceNameToClaimsOnly(t *testing.T) {
	claims := []resourcev1beta1.ResourceClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "selected-claim",
				Namespace: "default",
				Labels:    map[string]string{"app": "selected"},
			},
			Spec: resourcev1beta1.ResourceClaimSpec{
				Devices: resourcev1beta1.DeviceClaim{
					Requests: []resourcev1beta1.DeviceRequest{
						{Name: "accelerator", DeviceClassName: "gpu"},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "other-claim",
				Namespace: "default",
				Labels:    map[string]string{"app": "selected"},
			},
			Spec: resourcev1beta1.ResourceClaimSpec{
				Devices: resourcev1beta1.DeviceClaim{
					Requests: []resourcev1beta1.DeviceRequest{
						{Name: "accelerator", DeviceClassName: "missing-class"},
					},
				},
			},
		},
	}
	client := fake.NewSimpleClientset(&resourcev1beta1.DeviceClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "gpu",
			Labels: map[string]string{"tier": "platform"},
		},
	})
	// The fake client ignores field selectors, so emulate API server filtering.
	client.PrependReactor("list", "resourceclaims", func(action k8stesting.Action) (bool, runtime.Object, error) {
		listAction, ok := action.(k8stesting.ListAction)
		if !ok {
			return false, nil, nil
		}
		fieldSelector := listAction.GetListRestrictions().Fields
		out := &resourcev1beta1.ResourceClaimList{}
		for _, claim := range claims {
			if fieldSelector == nil || fieldSelector.Matches(fields.Set{"metadata.name": claim.Name}) {
				out.Items = append(out.Items, claim)
			}
		}
		return true, out, nil
	})
	config := common.Analyzer{
		Client:        &kubernetes.Client{Client: client},
		Context:       context.Background(),
		Namespace:     "default",
		LabelSelector: "app=selected",
		ResourceName:  "selected-claim",
	}

	results, err := (ResourceClaimAnalyzer{}).Analyze(config)
	require.NoError(t, err)
	require.Empty(t, results)

	restrictions := make(map[string]k8stesting.ListRestrictions)
	for _, action := range client.Actions() {
		if listAction, ok := action.(k8stesting.ListAction); ok {
			restrictions[action.GetResource().Resource] = listAction.GetListRestrictions()
		}
	}
	require.Contains(t, restrictions, "resourceclaims")
	require.Equal(t, "app=selected", restrictions["resourceclaims"].Labels.String())
	require.Equal(t, "metadata.name=selected-claim", restrictions["resourceclaims"].Fields.String())
	require.Contains(t, restrictions, "deviceclasses")
	require.True(t, restrictions["deviceclasses"].Labels.Empty())
	require.True(t, restrictions["deviceclasses"].Fields.Empty())
}
