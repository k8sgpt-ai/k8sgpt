// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	monitoring "github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/gmp/apis/monitoring"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	Version = "v1"
)

var (
	// SchemeBuilder initializes a scheme builder.
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	// AddToScheme is a global function that registers this API group & version to a scheme.
	AddToScheme = SchemeBuilder.AddToScheme
	// SchemeGroupVersion is group version used to register these objects.
	SchemeGroupVersion = schema.GroupVersion{Group: monitoring.GroupName, Version: Version}
)

// Kind takes an unqualified kind and returns back a Group qualified GroupKind.
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource.
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// PodMonitoringResource returns a PodMonitoring GroupVersionResource.
// This can be used to enforce API types.
func PodMonitoringResource() metav1.GroupVersionResource {
	return metav1.GroupVersionResource{
		Group:    monitoring.GroupName,
		Version:  Version,
		Resource: "podmonitorings",
	}
}

// ClusterPodMonitoringResource returns a ClusterPodMonitoring GroupVersionResource.
// This can be used to enforce API types.
func ClusterPodMonitoringResource() metav1.GroupVersionResource {
	return metav1.GroupVersionResource{
		Group:    monitoring.GroupName,
		Version:  Version,
		Resource: "clusterpodmonitorings",
	}
}

// OperatorConfigResource returns a OperatorConfig GroupVersionResource.
// This can be used to enforce API types.
func OperatorConfigResource() metav1.GroupVersionResource {
	return metav1.GroupVersionResource{
		Group:    monitoring.GroupName,
		Version:  Version,
		Resource: "operatorconfigs",
	}
}

// GlobalRulesResource returns a GlobalRules GroupVersionResource.
// This can be used to enforce API types.
func GlobalRulesResource() metav1.GroupVersionResource {
	return metav1.GroupVersionResource{
		Group:    monitoring.GroupName,
		Version:  Version,
		Resource: "globalrules",
	}
}

// ClusterRulesResource returns a ClusterRules GroupVersionResource.
// This can be used to enforce API types.
func ClusterRulesResource() metav1.GroupVersionResource {
	return metav1.GroupVersionResource{
		Group:    monitoring.GroupName,
		Version:  Version,
		Resource: "clusterrules",
	}
}

// RulesResource returns a Rules GroupVersionResource.
// This can be used to enforce API types.
func RulesResource() metav1.GroupVersionResource {
	return metav1.GroupVersionResource{
		Group:    monitoring.GroupName,
		Version:  Version,
		Resource: "rules",
	}
}

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&PodMonitoring{},
		&PodMonitoringList{},
		&ClusterPodMonitoring{},
		&ClusterPodMonitoringList{},
		&Rules{},
		&RulesList{},
		&ClusterRules{},
		&ClusterRulesList{},
		&GlobalRules{},
		&GlobalRulesList{},
		&OperatorConfig{},
		&OperatorConfigList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
