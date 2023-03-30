package analyzer

import (
	"context"
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AnalyzeIngress(ctx context.Context, config *AnalysisConfiguration, client *kubernetes.Client, aiClient ai.IAI,
	analysisResults *[]Analysis) error {

	list, err := client.GetClient().NetworkingV1().Ingresses(config.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	var preAnalysis = map[string]PreAnalysis{}

	for _, ing := range list.Items {
		var failures []string
		// loop over rules
		for _, rule := range ing.Spec.Rules {
			// loop over paths
			for _, path := range rule.HTTP.Paths {
				_, err := client.GetClient().CoreV1().Services(ing.Namespace).Get(ctx, path.Backend.Service.Name, metav1.GetOptions{})
				if err != nil {
					failures = append(failures, fmt.Sprintf("Ingress uses the service %s/%s which does not exist.", ing.Namespace, path.Backend.Service.Name))
				}
			}
		}

		for _, tls := range ing.Spec.TLS {
			_, err := client.GetClient().CoreV1().Secrets(ing.Namespace).Get(ctx, tls.SecretName, metav1.GetOptions{})
			if err != nil {
				failures = append(failures, fmt.Sprintf("Ingress uses the secret %s/%s as a TLS certificate which does not exist.", ing.Namespace, tls.SecretName))
			}
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", ing.Namespace, ing.Name)] = PreAnalysis{
				Ingress:        ing,
				FailureDetails: failures,
			}
		}

	}

	for key, value := range preAnalysis {
		var currentAnalysis = Analysis{
			Kind:  "Ingress",
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, _ := util.GetParent(client, value.Ingress.ObjectMeta)
		currentAnalysis.ParentObject = parent
		*analysisResults = append(*analysisResults, currentAnalysis)
	}

	return nil
}
