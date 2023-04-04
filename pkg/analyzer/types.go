package analyzer

import (
	"context"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	autov1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
)

type Analyzer struct {
	Client      *kubernetes.Client
	Context     context.Context
	Namespace   string
	AIClient    ai.IAI
	PreAnalysis map[string]PreAnalysis
	Results     []Result
}

type PreAnalysis struct {
	Pod                      v1.Pod
	FailureDetails           []string
	ReplicaSet               appsv1.ReplicaSet
	PersistentVolumeClaim    v1.PersistentVolumeClaim
	Endpoint                 v1.Endpoints
	Ingress                  networkv1.Ingress
	HorizontalPodAutoscalers autov1.HorizontalPodAutoscaler
	PodDisruptionBudget      policyv1.PodDisruptionBudget
}

type Result struct {
	Kind         string   `json:"kind"`
	Name         string   `json:"name"`
	Error        []string `json:"error"`
	Details      string   `json:"details"`
	ParentObject string   `json:"parentObject"`
}
