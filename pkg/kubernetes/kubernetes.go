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
	policyreport "github.com/kyverno/policy-reporter-kyverno-plugin/pkg/crd/api/policyreport/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	gtwapi "sigs.k8s.io/gateway-api/apis/v1"
)

var (
	installClientGo     = clientgoscheme.AddToScheme
	installGatewayAPI   = gtwapi.Install
	installPolicyReport = policyreport.AddToScheme
)

// newScheme builds the scheme for one client: the built-in Kubernetes types
// plus the CRD types the analyzers need.
//
// It is deliberately private to the client rather than controller-runtime's
// package-global default, which ctrl.New falls back to whenever Options.Scheme
// is nil. Installing onto that default made every NewClient a writer of state
// that another client's analyzers were concurrently reading, so the data race
// behind issue #1063 survived moving registration out of Analyze() — it just
// moved from two analyzers racing each other to client construction racing
// analysis. A scheme no one else holds cannot be written from under them.
func newScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	for _, install := range []func(*runtime.Scheme) error{
		installClientGo,
		installGatewayAPI,
		installPolicyReport,
	} {
		if err := install(scheme); err != nil {
			return nil, err
		}
	}
	return scheme, nil
}

func (c *Client) GetConfig() *rest.Config {
	return c.Config
}

func (c *Client) GetClient() kubernetes.Interface {
	return c.Client
}

func (c *Client) GetCtrlClient() ctrl.Client {
	return c.CtrlClient
}

func (c *Client) GetDynamicClient() dynamic.Interface {
	return c.DynamicClient
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

	// The scheme is fully populated before the client is built, so nothing
	// writes it once it is reachable from a client the analyzers are using.
	// Analyzers run concurrently against one shared client, and registering
	// types on that hot path races on the scheme's internal maps (issue #1063).
	scheme, err := newScheme()
	if err != nil {
		return nil, err
	}

	ctrlClient, err := ctrl.New(config, ctrl.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	serverVersion, err := clientSet.ServerVersion()
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client:        clientSet,
		CtrlClient:    ctrlClient,
		Config:        config,
		ServerVersion: serverVersion,
		DynamicClient: dynamicClient,
	}, nil
}
