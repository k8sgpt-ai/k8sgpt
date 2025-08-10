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

package common

import (
	"context"
	"time"

	openapi_v2 "github.com/google/gnostic/openapiv2"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	keda "github.com/kedacore/keda/v2/apis/keda/v1alpha1"
	kyverno "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/crd/api/policyreport/v1alpha2"
	regv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	autov2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gtwapi "sigs.k8s.io/gateway-api/apis/v1"
)

type IAnalyzer interface {
	Analyze(analysis Analyzer) ([]Result, error)
}

type Analyzer struct {
	Client        *kubernetes.Client
	Context       context.Context
	Namespace     string
	LabelSelector string
	AIClient      ai.IAI
	PreAnalysis   map[string]PreAnalysis
	Results       []Result
	OpenapiSchema *openapi_v2.Document
}

type PreAnalysis struct {
	Pod                      v1.Pod
	FailureDetails           []Failure
	Deployment               appsv1.Deployment
	ReplicaSet               appsv1.ReplicaSet
	PersistentVolumeClaim    v1.PersistentVolumeClaim
	Endpoint                 v1.Endpoints
	Ingress                  networkv1.Ingress
	HorizontalPodAutoscalers autov2.HorizontalPodAutoscaler
	PodDisruptionBudget      policyv1.PodDisruptionBudget
	StatefulSet              appsv1.StatefulSet
	NetworkPolicy            networkv1.NetworkPolicy
	Node                     v1.Node
	ValidatingWebhook        regv1.ValidatingWebhookConfiguration
	MutatingWebhook          regv1.MutatingWebhookConfiguration
	GatewayClass             gtwapi.GatewayClass
	Gateway                  gtwapi.Gateway
	HTTPRoute                gtwapi.HTTPRoute
	// Integrations
	ScaledObject               keda.ScaledObject
	KyvernoPolicyReport        kyverno.PolicyReport
	KyvernoClusterPolicyReport kyverno.ClusterPolicyReport
	Catalog                    ClusterCatalog
	Extension                  ClusterExtension
}

type Result struct {
	Kind         string    `json:"kind"`
	Name         string    `json:"name"`
	Error        []Failure `json:"error"`
	Details      string    `json:"details"`
	ParentObject string    `json:"parentObject"`
}

type AnalysisStats struct {
	Analyzer     string        `json:"analyzer"`
	DurationTime time.Duration `json:"durationTime"`
}

type Failure struct {
	Text          string
	KubernetesDoc string
	Sensitive     []Sensitive
}

type Sensitive struct {
	Unmasked string
	Masked   string
}

type (
	SourceType                  string
	AvailabilityMode            string
	UpgradeConstraintPolicy     string
	CRDUpgradeSafetyEnforcement string
)

type ClusterCatalog struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ClusterCatalogSpec   `json:"spec"`
	Status            ClusterCatalogStatus `json:"status,omitempty"`
}
type ClusterCatalogList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ClusterCatalog `json:"items"`
}

type ClusterCatalogSpec struct {
	Source CatalogSource `json:"source"`

	Priority int32 `json:"priority"`

	AvailabilityMode AvailabilityMode `json:"availabilityMode,omitempty"`
}

type ClusterCatalogStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	ResolvedSource *ResolvedCatalogSource `json:"resolvedSource,omitempty"`

	URLs         *ClusterCatalogURLs `json:"urls,omitempty"`
	LastUnpacked *metav1.Time        `json:"lastUnpacked,omitempty"`
}

type ClusterCatalogURLs struct {
	Base string `json:"base"`
}
type CatalogSource struct {
	Type  SourceType   `json:"type"`
	Image *ImageSource `json:"image,omitempty"`
}
type ResolvedCatalogSource struct {
	Type  SourceType           `json:"type"`
	Image *ResolvedImageSource `json:"image"`
}
type ResolvedImageSource struct {
	Ref string `json:"ref"`
}

type ImageSource struct {
	Ref                 string `json:"ref"`
	PollIntervalMinutes *int   `json:"pollIntervalMinutes,omitempty"`
}

type ClusterExtension struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ClusterExtensionSpec   `json:"spec,omitempty"`
	Status            ClusterExtensionStatus `json:"status,omitempty"`
}

type ClusterExtensionSpec struct {
	Namespace      string                         `json:"namespace"`
	ServiceAccount ServiceAccountReference        `json:"serviceAccount"`
	Source         SourceConfig                   `json:"source"`
	Install        *ClusterExtensionInstallConfig `json:"install,omitempty"`
}

type ClusterExtensionInstallConfig struct {
	Preflight *PreflightConfig `json:"preflight,omitempty"`
}

type PreflightConfig struct {
	CRDUpgradeSafety *CRDUpgradeSafetyPreflightConfig `json:"crdUpgradeSafety"`
}

type CRDUpgradeSafetyPreflightConfig struct {
	Enforcement CRDUpgradeSafetyEnforcement `json:"enforcement"`
}

type ServiceAccountReference struct {
	Name string `json:"name"`
}

type SourceConfig struct {
	SourceType string         `json:"sourceType"`
	Catalog    *CatalogFilter `json:"catalog,omitempty"`
}

type CatalogFilter struct {
	PackageName             string                  `json:"packageName"`
	Version                 string                  `json:"version,omitempty"`
	Channels                []string                `json:"channels,omitempty"`
	Selector                *metav1.LabelSelector   `json:"selector,omitempty"`
	UpgradeConstraintPolicy UpgradeConstraintPolicy `json:"upgradeConstraintPolicy,omitempty"`
}

type ClusterExtensionStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	Install *ClusterExtensionInstallStatus `json:"install,omitempty"`
}

type ClusterExtensionInstallStatus struct {
	Bundle BundleMetadata `json:"bundle"`
}

type BundleMetadata struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
