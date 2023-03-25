package analyzer

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

type PreAnalysis struct {
	Pod            v1.Pod
	FailureDetails []string
	ReplicaSet     appsv1.ReplicaSet
}

type Analysis struct {
	Kind         string `json:"kind"`
	Name         string `json:"name"`
	Error        string `json:"error"`
	Details      string `json:"details"`
	ParentObject string `json:"parentObject"`
}
