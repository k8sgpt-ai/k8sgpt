package kubernetes

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	client *kubernetes.Clientset
}

func (c *Client) GetClient() *kubernetes.Clientset {
	return c.client
}

func NewClient(masterURL string, kubeconfig string, context string) (*Client, error) {

	config, err := rest.InClusterConfig()
	if err != nil {
		loaderRules := clientcmd.NewDefaultClientConfigLoadingRules()

		if kubeconfig != "" {
			loaderRules.Precedence = []string{kubeconfig}
		}

		configOverrides := &clientcmd.ConfigOverrides{}

		if masterURL != "" {
			configOverrides.ClusterInfo.Server = masterURL
		}

		if context != "" {
			configOverrides.CurrentContext = context
		}

		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loaderRules, configOverrides)
		config, err = kubeConfig.ClientConfig()
		if err != nil {
			return nil, err
		}
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Client{
		client: clientSet,
	}, nil
}
