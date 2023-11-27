/*
Copyright 2023 The K8sGPT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubernetes

import (
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *Client) GetConfig() *rest.Config {
	return c.Config
}

func (c *Client) GetClient() kubernetes.Interface {
	return c.Client
}

func (c *Client) GetCtrlClient() ctrl.Client {
	return c.CtrlClient
}

func NewClient(kubecontext string, kubeconfig string) (*Client, error) {
	var config *rest.Config
	config, err := rest.InClusterConfig()
	if kubeconfig != "" || err != nil {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

		if kubeconfig != "" {
			loadingRules.ExplicitPath = kubeconfig
		}

		clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			loadingRules,
			&clientcmd.ConfigOverrides{
				CurrentContext: kubecontext,
			})
		// create the clientset
		config, err = clientConfig.ClientConfig()
		if err != nil {
			return nil, err
		}
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	ctrlClient, err := ctrl.New(config, ctrl.Options{})
	if err != nil {
		return nil, err
	}

	serverVersion, err := clientSet.ServerVersion()
	if err != nil {
		return nil, err
	}

	return &Client{
		Client:        clientSet,
		CtrlClient:    ctrlClient,
		Config:        config,
		ServerVersion: serverVersion,
	}, nil
}
