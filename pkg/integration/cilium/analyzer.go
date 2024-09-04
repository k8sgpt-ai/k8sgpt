package cilium

import (
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	v1 "k8s.io/api/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

type CiliumAnalyzer struct {
}

func (c CiliumAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	results := make([]common.Result, 0)
	namespace := "kube-system"
	if a.Namespace != "" {
		namespace = a.Namespace
	}

	ciliumAgentDsResult, err := analyzeDaemonset(a, namespace, "cilium")
	if err != nil {
		return nil, err
	}
	results = append(results, ciliumAgentDsResult...)

	ciliumEnvoyDsResult, err := analyzeDaemonset(a, namespace, "cilium-envoy")
	if err != nil {
		return nil, err
	}
	results = append(results, ciliumEnvoyDsResult...)

	ciliumOperatorDeployResult, err := analyzeDeployment(a, namespace, "cilium-operator")
	if err != nil {
		return nil, err
	}
	results = append(results, ciliumOperatorDeployResult...)

	hubbleRelayDeployResult, err := analyzeDeployment(a, namespace, "hubble-relay")
	if err != nil {
		return nil, err
	}
	results = append(results, hubbleRelayDeployResult...)

	hubbleUiDeployResult, err := analyzeDeployment(a, namespace, "hubble-ui")
	if err != nil {
		return nil, err
	}
	results = append(results, hubbleUiDeployResult...)

	ciliumClusterMeshDeployResult, err := analyzeDeployment(a, namespace, "clustermesh-apiserver")
	if err != nil {
		return nil, err
	}
	results = append(results, ciliumClusterMeshDeployResult...)

	return results, nil
}

func analyzeDaemonset(a common.Analyzer, namespace string, dsName string) ([]common.Result, error) {
	var preAnalysis = map[string]common.PreAnalysis{}
	var failures []common.Failure
	client := a.Client.CtrlClient

	dsObj := &v1.DaemonSet{}
	err := client.Get(a.Context, ctrl.ObjectKey{
		Namespace: namespace,
		Name:      dsName,
	}, dsObj)
	if err != nil {
		return nil, err
	}

	notReady := int(dsObj.Status.DesiredNumberScheduled) - int(dsObj.Status.NumberReady)
	if notReady > 0 {
		failures = append(failures, common.Failure{
			Text:      fmt.Sprintf("%d pods of DaemonSet %s are not ready", notReady, dsName),
			Sensitive: []common.Sensitive{},
		})
	}

	if unavailable := int(dsObj.Status.NumberUnavailable) - notReady; unavailable > 0 {
		failures = append(failures, common.Failure{
			Text:      fmt.Sprintf("%d pods of DaemonSet %s are not available", unavailable, dsName),
			Sensitive: []common.Sensitive{},
		})
	}

	if dsObj.Status.UpdatedNumberScheduled < dsObj.Status.DesiredNumberScheduled {
		failures = append(failures, common.Failure{
			Text:      fmt.Sprintf("DaemonSet %s is rolling out - %d out of %d pods updated", dsName, dsObj.Status.UpdatedNumberScheduled, dsObj.Status.DesiredNumberScheduled),
			Sensitive: []common.Sensitive{},
		})
	}

	if len(failures) > 0 {
		preAnalysis[fmt.Sprintf("%s/%s", dsObj.Namespace,
			dsObj.Name)] = common.PreAnalysis{
			DaemonSet:      *dsObj,
			FailureDetails: failures,
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  "DaemonSet",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.DaemonSet.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}

func analyzeDeployment(a common.Analyzer, namespace string, deployName string) ([]common.Result, error) {
	var preAnalysis = map[string]common.PreAnalysis{}
	var failures []common.Failure
	client := a.Client.CtrlClient

	deployObj := &v1.Deployment{}
	err := client.Get(a.Context, ctrl.ObjectKey{
		Namespace: namespace,
		Name:      deployName,
	}, deployObj)
	if err != nil {
		return nil, err
	}

	notReady := int(deployObj.Status.Replicas) - int(deployObj.Status.ReadyReplicas)
	if notReady > 0 {
		failures = append(failures, common.Failure{
			Text:      fmt.Sprintf("%d pods of Deployment %s are not ready", notReady, deployName),
			Sensitive: []common.Sensitive{},
		})
	}

	if unavailable := int(deployObj.Status.UnavailableReplicas) - notReady; unavailable > 0 {
		failures = append(failures, common.Failure{
			Text:      fmt.Sprintf("%d pods of Deployment %s are not available", unavailable, deployName),
			Sensitive: []common.Sensitive{},
		})
	}

	if deployObj.Status.UpdatedReplicas < deployObj.Status.Replicas {
		failures = append(failures, common.Failure{
			Text:      fmt.Sprintf("Deployment %s is rolling out - %d out of %d pods updated", deployName, deployObj.Status.UpdatedReplicas, deployObj.Status.Replicas),
			Sensitive: []common.Sensitive{},
		})
	}

	if len(failures) > 0 {
		preAnalysis[fmt.Sprintf("%s/%s", deployObj.Namespace,
			deployObj.Name)] = common.PreAnalysis{
			Deployment:     *deployObj,
			FailureDetails: failures,
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  "Deployment",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.Deployment.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}
