package kubernetes

import (

	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	Client kubernetes.Interface
}

func (c *Client) GetClient() kubernetes.Interface {
	return c.Client
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

	return &Client{
		Client: clientSet,
	}, nil
}
