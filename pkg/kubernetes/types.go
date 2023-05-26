package kubernetes

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	Client        kubernetes.Interface
	RestClient    rest.Interface
	Config        *rest.Config
	ServerVersion *version.Info
}

type K8sApiReference struct {
	ApiVersion schema.GroupVersion
	Kind       string
	// Property   string
	Discovery discovery.DiscoveryInterface
}
