package analyzer

import (
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
)

type AnalysisConfiguration struct {
	Namespace string
	NoCache   bool
	Explain   bool
}

type PreAnalysis struct {
	Pod                      v1.Pod
	FailureDetails           []string
	ReplicaSet               appsv1.ReplicaSet
	PersistentVolumeClaim    v1.PersistentVolumeClaim
	Endpoint                 v1.Endpoints
	Ingress                  networkingv1.Ingress
	HorizontalPodAutoscalers autoscalingv1.HorizontalPodAutoscaler
	PodDisruptionBudget      policyv1.PodDisruptionBudget
}

type Analysis struct {
	Kind         string   `json:"kind"`
	Name         string   `json:"name"`
	Error        []string `json:"error"`
	Details      string   `json:"details"`
	ParentObject string   `json:"parentObject"`
}
