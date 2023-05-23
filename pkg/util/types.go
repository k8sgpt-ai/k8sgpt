package util

import "k8s.io/apimachinery/pkg/version"

type K8sApiReference struct {
	ApiVersion    string
	Kind          string
	ServerVersion *version.Info
}
