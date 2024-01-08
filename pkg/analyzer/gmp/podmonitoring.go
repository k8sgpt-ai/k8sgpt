package gmp

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/fatih/color"
	monitoring "github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/gmp/apis/monitoring"
	v1 "github.com/k8sgpt-ai/k8sgpt/pkg/analyzer/gmp/apis/monitoring/v1"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

func newCRClient(coreClient *kubernetes.Client, crdInfo *schema.GroupVersion) (*rest.RESTClient, error) {
	config := *coreClient.GetConfig() // Shallow copy.
	config.ContentConfig.GroupVersion = crdInfo
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	config.UserAgent = rest.DefaultKubernetesUserAgent()
	return rest.UnversionedRESTClientFor(&config)
}

type podMonitoringAnalysis struct {
	cfg common.Analyzer

	clusterName string
	client      *rest.RESTClient
	promClient  promv1.API
}

func AnalyzePodMonitorings(a common.Analyzer) ([]common.Result, error) {
	ctx := context.TODO()
	client, err := newCRClient(a.Client, &schema.GroupVersion{Group: monitoring.GroupName, Version: "v1"})
	if err != nil {
		return nil, err
	}

	// Injected e.g. by: gcloud auth application-default login
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/monitoring.read")
	if err != nil {
		return nil, err
	}

	// TODO(bwplotka): Find a better way to find cluster and GCM project ID.
	// See: https://stackoverflow.com/a/62542265
	out, err := exec.Command("kubectl", "config", "view", "--minify", "-o", "jsonpath='{.clusters[].name}'").Output()
	if err != nil {
		return nil, fmt.Errorf("detecting cluster name from kubectl; is kubectl installed and it's context is set? Use 'gcloud container clusters get-credentials <cluster name> --zone <zone> --project <project ID> otherwise; got %w, out: %v", err, string(out))
	}

	gkeFullClusterName := strings.Trim(string(out), "'") // Should look something like gke_<project>_<zone>_<name>
	if !strings.HasPrefix(gkeFullClusterName, "gke_") {
		return nil, fmt.Errorf("local cluster config name does not indicate gke cluster, make sure the name is in the form of gke_<project>_<zone>_<name>; got %v", gkeFullClusterName)
	}

	s := strings.Split(gkeFullClusterName, "_")
	if len(s) != 4 {
		return nil, fmt.Errorf("local cluster config name is not in form of gke_<project>_<zone>_<name>; got %v", gkeFullClusterName)
	}

	gcmProjectID := s[1]
	if creds.ProjectID != "" {
		gcmProjectID = creds.ProjectID
	}

	cl, err := api.NewClient(api.Config{
		Address: fmt.Sprintf("https://monitoring.googleapis.com/v1/projects/%s/location/global/prometheus", gcmProjectID),
		Client:  oauth2.NewClient(ctx, creds.TokenSource),
	})
	if err != nil {
		return nil, err
	}
	promClient := promv1.NewAPI(cl)

	return (&podMonitoringAnalysis{
		cfg:         a,
		clusterName: s[3],
		client:      client,
		promClient:  promClient,
	}).Analyze(ctx)
}

func (a *podMonitoringAnalysis) Analyze(ctx context.Context) (res []common.Result, _ error) {
	const (
		kind  = "PodMonitoring"
		kinds = "PodMonitorings" // Used for API call, singular form seems to not work.
	)

	var list = &v1.PodMonitoringList{}
	if len(a.cfg.Resources[kind]) > 0 {
		for _, pm := range a.cfg.Resources[kind] {
			podMonitoring := v1.PodMonitoring{}
			if err := a.client.Get().Namespace(a.cfg.Namespace).Resource(kinds).Name(pm).Do(ctx).Into(&podMonitoring); err != nil {
				return nil, fmt.Errorf("get pod: %w", err)
			}
			list.Items = append(list.Items, podMonitoring)
		}
	} else {
		req := a.client.Get().Resource(kinds)
		if a.cfg.Namespace != "" {
			req.Namespace(a.cfg.Namespace)
		}
		if err := req.Do(ctx).Into(list); err != nil {
			return nil, fmt.Errorf("get related pods: %w", err)
		}
	}

	a.debugf("", "Found %v PodMonitoring resources to analyze for namespace %q", len(list.Items), a.cfg.Namespace)
	for _, pm := range list.Items {
		r, err := a.analyzePodMonitoring(ctx, pm)
		if err != nil {
			return nil, err
		}
		if len(r.Error) > 0 {
			res = append(res, r)
		}
	}
	return res, nil
}

func (a *podMonitoringAnalysis) debugf(resource string, f string, args ...any) {
	if !a.cfg.Verbose {
		return
	}
	if resource != "" {
		f = fmt.Sprintf("resource=%v %v", resource, f)
	}
	color.Blue("[PodMonitoring] Debug: "+f, args...)
}

func (a *podMonitoringAnalysis) errorf(resource string, f string, args ...any) {
	if !a.cfg.Verbose {
		return
	}
	if resource != "" {
		f = fmt.Sprintf("resource=%v %v", resource, f)
	}
	color.Red("[PodMonitoring] Error: "+f, args...)
}

func (a *podMonitoringAnalysis) analyzePodMonitoring(ctx context.Context, pm v1.PodMonitoring) (common.Result, error) {
	r := common.Result{
		Kind:         "PodMonitoring",
		Name:         pm.Namespace + "/" + pm.Name,
		ParentObject: pm.GetKey(),
	}

	a.debugf(pm.GetKey(), "analysing")
	instancesPerPod := len(pm.Spec.Endpoints)

	if len(pm.Status.EndpointStatuses) > 0 {
		// TODO(bwplotka): Check statuses if Target Status feature is enabled.
		a.debugf(pm.GetKey(), "target status feature seems enabled, got %v", pm.Status.EndpointStatuses)
	}

	var (
		notEnoughEndpointsErrByPod = map[string]error{}
		nothingFoundErr            error
	)
	{
		// Is it really a problem? Query GCM first.
		// TODO(bwplotka): Do in bulk for all podmonitoring CRs?
		promql := fmt.Sprintf(`sum(up{job="%s", namespace="%s", cluster="%s"}) by (pod)`, pm.Name, pm.Namespace, a.clusterName)
		if a.cfg.Verbose {
			a.debugf(pm.GetKey(), "querying GCM with %v", promql)
		}
		val, warns, err := a.promClient.Query(ctx, promql, time.Now())
		if err != nil {
			return r, fmt.Errorf("querying GCM: %w", err)
		}
		if len(warns) > 0 {
			color.Yellow("warnings from GCM %v", warns) // TODO(bwplotka): Format.
		}

		switch val.Type() {
		case model.ValVector:
			v := val.(model.Vector)
			if len(v) > 0 {
				for _, perPod := range v {
					if int(perPod.Value) < instancesPerPod {
						notEnoughEndpointsErrByPod[string(perPod.Metric["pod"])] = fmt.Errorf("PodMonitoring.Spec.Endpoints suggests %v endpoints per pod should be monitored, found sucessful scrapes only for %v targets for pod %v in GCM", instancesPerPod, perPod.Value, perPod.Metric["pod"])
						a.errorf(pm.GetKey(), notEnoughEndpointsErrByPod[string(perPod.Metric["pod"])].Error())
						continue
					}

					if int(perPod.Value) > instancesPerPod {
						color.Red("resource=%v found more targets than expected?")
					}
				}
				a.debugf(pm.GetKey(), "found following pods scraped:\n%v\n", v.String())
			} else {
				nothingFoundErr = fmt.Errorf("PodMonitoring.Spec.Endpoints suggests %v endpoints per pod should be monitored, found no sucessful scrapes for this job in GCM", instancesPerPod)
				a.errorf(pm.GetKey(), nothingFoundErr.Error())
			}
		default:
			return r, fmt.Errorf("unexpected PromQL response %v", val.Type())
		}

		if nothingFoundErr == nil && len(notEnoughEndpointsErrByPod) == 0 {
			a.debugf(pm.GetKey(), "seems to work fine")
			return r, nil
		}
	}
	// At this point we know we don't have expected number of pods/containers scraped.

	var problematicPod corev1.Pod
	{
		// Is anything selected, check first selected pod.
		selectorLabels, err := metav1.LabelSelectorAsMap(&pm.Spec.Selector)
		if err != nil {
			return r, err
		}
		pods, err := a.cfg.Client.GetClient().CoreV1().Pods(pm.Namespace).List(
			ctx,
			metav1.ListOptions{LabelSelector: labels.SelectorFromSet(selectorLabels).String()},
		)
		if err != nil {
			return r, fmt.Errorf("get related pods: %w", err)
		}
		if len(pods.Items) == 0 {
			// Might be race or different case.
			if nothingFoundErr == nil {
				for _, err := range notEnoughEndpointsErrByPod {
					nothingFoundErr = err
				}
				if nothingFoundErr == nil {
					nothingFoundErr = fmt.Errorf("PodMonitoring.Spec.Endpoints suggests %v endpoints per pod should be monitored, found no sucessful scrapes for this job in GCM", instancesPerPod)
				}
			}
			// TODO(bwplotka): Check all objects, to tell if there is any *sets selected for it?
			r.Error = append(r.Error, newFailure(
				nothingFoundErr,
				fmt.Sprintf("PodMonitoring.Spec.Selector selects %v; However no pods can be found for this selector in namespace %q", selectorLabels, pm.Namespace),
				"",
				nil,
			))
			// NextStepsText: fmt.Sprintf("Check the underlying Deployment, StatefulSet, ReplicaSet or DaemonSet?"), Let's see, maybe AI will be better at this.
			return r, nil
		}
		a.debugf(pm.GetKey(), "selects %v pods in %s namespace", len(pods.Items), pm.Namespace)
		problematicPod = pods.Items[0]
	}

	// TODO(bwplotka): Check other pods? Right now we assume all pods have similar problem, but it might be.. just tmp?
	// Check the pod.
	if problematicPod.Status.Phase != corev1.PodRunning {
		r.Error = append(r.Error, newFailure(
			nothingFoundErr,
			fmt.Sprintf("PodMonitoring.Spec.Selector selects pod %q in namespace %q, however that pod is not running", problematicPod.Name, problematicPod.Namespace),
			"",
			// TODO(bwplotka): Run Pod analyzer (cycling dependency issue, so yolo feature for now).
			[]common.Analyzer{func() common.Analyzer {
				cfg := a.cfg
				cfg.Resources = map[string][]string{
					"Pod": {problematicPod.Name},
				}
				return cfg
			}()},
		))
		return r, nil
	}

	var failureErr = nothingFoundErr
	// It's running, do we get any metric from it? If yes, try to match what endpoint is problematic.
	var problematicEndpoints []v1.ScrapeEndpoint
	if notEnoughEndpointsErrByPod[problematicPod.Name] != nil {
		failureErr = notEnoughEndpointsErrByPod[problematicPod.Name]
		// Some endpoints return metrics, let's see which endpoints.
		// TODO(bwplotka): Find those, for now yolo.
		problematicEndpoints = pm.Spec.Endpoints
	} else {
		// No endpoint works.
		problematicEndpoints = pm.Spec.Endpoints
	}

	for _, endpoint := range problematicEndpoints {
		failureAdditionalContext := a.analyzePodMonitoringEndpointPort(pm, endpoint, problematicPod)
		if failureAdditionalContext != "" {
			r.Error = append(r.Error, newFailure(
				failureErr,
				fmt.Sprintf("PodMonitoring.Spec.Selector selects pod %q in namespace %q%s", problematicPod.Name, problematicPod.Namespace, failureAdditionalContext),
				"",
				nil,
			))
			return r, nil
		}

		if endpoint.Path == "" {
			// Yolo.
			endpoint.Path = "/metrics"
		}
		// TODO(bwplotka): Check scheme, path, relabelling?
		r.Error = append(r.Error, newFailure(
			failureErr,
			"Pods are running, PodMonitoring selects them correctly. Specified Port is exposed.", // Print pod vs PodMonioring and ask for the follow up?
			fmt.Sprintf(`1. If you changed configuration recently, it takes a ~minute for samples to propagate to the GCM.
2. Check if the pod %v port %v actually serve Prometheus metrics on HTTP %s path.
3. Check if the Scheme match, maybe the endpoint is behind TLS?
4. Check if GMP is enabled.`, problematicPod.Name, endpoint.Port, endpoint.Path),
			nil,
		))
	}
	return r, nil
}

func (a *podMonitoringAnalysis) analyzePodMonitoringEndpointPort(pm v1.PodMonitoring, e v1.ScrapeEndpoint, pod corev1.Pod) (failureAdditionalContext string) {
	// Check port.
	tcpPortsByContainer := map[string][]string{}
	for _, c := range pod.Spec.Containers {
		for _, p := range c.Ports {
			if p.Protocol != corev1.ProtocolTCP {
				continue
			}

			if p.Name != "" && e.Port.String() == p.Name {
				// Match by name, this part of endpoint is fine.
				a.debugf(pm.GetKey(), "port name %v match", p.Name)
				return ""
			}
			// Ignoring p.HostPort for now.
			if e.Port.String() == strconv.Itoa(int(p.ContainerPort)) {
				// Match by port number, this part of endpoint is fine.
				a.debugf(pm.GetKey(), "port %v match", p.ContainerPort)
				return ""
			}

			tcpPortsByContainer[c.Name] = append(tcpPortsByContainer[c.Name], fmt.Sprintf("{name:%q, containerPort:%q}", p.Name, strconv.Itoa(int(p.ContainerPort))))
		}
	}

	return fmt.Sprintf(", however none of containers' ports match PodMonitoring.Spec.ScrapeEndpoint.Port %q; the pod has following TCP ports:%v", e.Port.String(), func() string {
		var msg string
		for c, ports := range tcpPortsByContainer {
			msg += fmt.Sprintf("\n    Container %q: %v", c, strings.Join(ports, ","))
		}
		return msg
	}())
}

var prompt = template.Must(template.New("default").Parse(
	`Imagine you are a Google Managed Prometheus ("GMP") product expert helping Google Kubernetes Engine customer. Don't mention Prometheus Operator. Be specific. Simplify the following Kubernetes error message delimited by triple dashes is written in --- {{ .Failure.Text }} ---. Additional context for this problem delimited by triple dashes is written in --- {{ .Failure.AdditionalContextText }} ---.

Provide the most possible solution in a step by step style in no more than 400 characters. Use {{ .Language }} language in your response. 

{{ if .Failure.NextStepsText }}Include the following solution delimited by triple dashes is written in --- {{ .Failure.NextStepsText }} ---.{{end}}
Write the output in the following format:

Error: {Explain error here}

Solution: {Step by step solution here}

Additional Resources: {Additional resources mentioning Google Managed Prometheus product }.`, // TODO(bwplotka): Play with next steps.
))

func newFailure(err error, furtherFindings string, potentialNextStep string, scheduledAnalysis []common.Analyzer, questions ...string) common.Failure {
	return common.Failure{
		Text:                  err.Error(),
		AdditionalContextText: furtherFindings,
		NextStepsText:         potentialNextStep,
		CustomPromptTemplate:  prompt,
		ScheduledAnalysis:     scheduledAnalysis,
		UsefulQuestions:       questions,
	}
}
