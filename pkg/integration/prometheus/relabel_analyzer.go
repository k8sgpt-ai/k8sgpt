package prometheus

import (
	"fmt"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	discoverykube "github.com/prometheus/prometheus/discovery/kubernetes"
	"gopkg.in/yaml.v2"
)

type RelabelAnalyzer struct {
}

func (r *RelabelAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	ctx := a.Context
	client := a.Client.GetClient()
	namespace := a.Namespace
	kind := ConfigRelabel

	podConfigs, err := findPrometheusPodConfigs(ctx, client, namespace)
	if err != nil {
		return nil, err
	}

	var preAnalysis = map[string]common.PreAnalysis{}
	for _, pc := range podConfigs {
		var failures []common.Failure
		pod := pc.pod

		// Check upstream validation.
		// The Prometheus configuration structs do not generally have validation
		// methods and embed their validation logic in the UnmarshalYAML methods.
		config, _ := unmarshalPromConfigBytes(pc.b)
		// Limit output for brevity.
		limit := 6
		i := 0
		for _, sc := range config.ScrapeConfigs {
			if i == limit {
				break
			}
			if sc == nil {
				continue
			}
			brc, _ := yaml.Marshal(sc.RelabelConfigs)
			var bsd []byte
			for _, cfg := range sc.ServiceDiscoveryConfigs {
				ks, ok := cfg.(*discoverykube.SDConfig)
				if !ok {
					continue
				}
				bsd, _ = yaml.Marshal(ks)
			}
			// Don't bother with relabel analysis if the scrape config
			// or service discovery config are empty.
			if len(brc) == 0 || len(bsd) == 0 {
				continue
			}
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf("job_name:\n%s\nrelabel_configs:\n%s\nkubernetes_sd_configs:\n%s\n", sc.JobName, string(brc), string(bsd)),
			})
			i++
		}

		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)] = common.PreAnalysis{
				Pod:            *pod,
				FailureDetails: failures,
			}
		}
	}

	for key, value := range preAnalysis {
		var currentAnalysis = common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}
		parent, _ := util.GetParent(a.Client, value.Pod.ObjectMeta)
		currentAnalysis.ParentObject = parent
		a.Results = append(a.Results, currentAnalysis)
	}

	return a.Results, nil
}
