package kubernetes

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	Client     kubernetes.Interface
	RestClient rest.Interface
}

func (c *Client) GetClient() kubernetes.Interface {
	return c.Client
}

func (c *Client) GetRestClient() rest.Interface {
	return c.RestClient
}

func NewClient(kubecontext string, kubeconfig string) (*Client, error) {

	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{
			CurrentContext: kubecontext,
		})
	// create the clientset
	c, err := config.ClientConfig()
	if err != nil {
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	restClient, err := rest.RESTClientFor(c)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client:     clientSet,
		RestClient: restClient,
	}, nil
}
