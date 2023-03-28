package analyzer

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

type AnalysisConfiguration struct {
	Namespace string
	NoCache   bool
	Explain   bool
}

type PreAnalysis struct {
	Pod                   v1.Pod
	FailureDetails        []string
	ReplicaSet            appsv1.ReplicaSet
	PersistentVolumeClaim v1.PersistentVolumeClaim
	Endpoint              v1.Endpoints
}

type Analysis struct {
	Kind         string   `json:"kind"`
	Name         string   `json:"name"`
	Error        []string `json:"error"`
	Details      string   `json:"details"`
	ParentObject string   `json:"parentObject"`
}
