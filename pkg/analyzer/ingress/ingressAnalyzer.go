package ingress

import (
	"fmt"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IngressAnalyzer struct{}

func (IngressAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {

	list, err := a.Client.GetClient().NetworkingV1().Ingresses(a.Namespace).List(a.Context, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, ing := range list.Items {
		var failures []string

		// get ingressClassName
		ingressClassName := ing.Spec.IngressClassName
		if ingressClassName == nil {
			ingClassValue := ing.Annotations["kubernetes.io/ingress.class"]
			if ingClassValue == "" {
				failures = append(failures, fmt.Sprintf("Ingress %s/%s does not specify an Ingress class.", ing.Namespace, ing.Name))
			} else {
				ingressClassName = &ingClassValue
			}
		}

		// check if ingressclass exist
		if ingressClassName != nil {
			_, err := a.Client.GetClient().NetworkingV1().IngressClasses().Get(a.Context, *ingressClassName, metav1.GetOptions{})
			if err != nil {
				failures = append(failures, fmt.Sprintf("Ingress uses the ingress class %s which does not exist.", *ingressClassName))
			}
		}

		// loop over rules
		for _, rule := range ing.Spec.Rules {
			// loop over paths
			for _, path := range rule.HTTP.Paths {
				_, err := a.Client.GetClient().CoreV1().Services(ing.Namespace).Get(a.Context, path.Backend.Service.Name, metav1.GetOptions{})
				if err != nil {
					failures = append(failures, fmt.Sprintf("Ingress uses the service %s/%s which does not exist.", ing.Namespace, path.Backend.Service.Name))
				}
			}
		}

		for _, tls := range ing.Spec.TLS {
			_, err := a.Client.GetClient().CoreV1().Secrets(ing.Namespace).Get(a.Context, tls.SecretName, metav1.GetOptions{})
			if err != nil {
				failures = append(failures, fmt.Sprintf("Ingress uses the secret %s/%s as a TLS certificate which does not exist.", ing.Namespace, tls.SecretName))
			}
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", ing.Namespace, ing.Name)] = common.PreAnalysis{
				Ingress:        ing,
				FailureDetails: failures,
			}
		}

	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  "Ingress",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(a.Client, value.Ingress.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}
