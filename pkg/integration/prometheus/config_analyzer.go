package prometheus

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	promconfig "github.com/prometheus/prometheus/config"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	prometheusContainerName     = "prometheus"
	configReloaderContainerName = "config-reloader"
	prometheusConfigFlag        = "--config.file="
	configReloaderConfigFlag    = "--config-file="
)

var prometheusPodLabels = map[string]string{
	"app":                    "prometheus",
	"app.kubernetes.io/name": "prometheus",
}

type ConfigAnalyzer struct {
}

// podConfig groups a specific pod with the Prometheus configuration and any
// other state used for informing the common.Result.
type podConfig struct {
	b   []byte
	pod *corev1.Pod
}

func (c *ConfigAnalyzer) Analyze(a common.Analyzer) ([]common.Result, error) {
	ctx := a.Context
	client := a.Client.GetClient()
	namespace := a.Namespace
	kind := ConfigValidate

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
		config, err := unmarshalPromConfigBytes(pc.b)
		if err != nil {
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf("error validating Prometheus YAML configuration: %s", err),
			})
		}
		_, err = yaml.Marshal(config)
		if err != nil {
			failures = append(failures, common.Failure{
				Text: fmt.Sprintf("error validating Prometheus struct configuration: %s", err),
			})
		}

		// Check for empty scrape config.
		if len(config.ScrapeConfigs) == 0 {
			failures = append(failures, common.Failure{
				Text: "no scrape configurations. Prometheus will not scrape any metrics.",
			})
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

func configKey(namespace string, volume *corev1.Volume) (string, error) {
	if volume.ConfigMap != nil {
		return fmt.Sprintf("configmap/%s/%s", namespace, volume.ConfigMap.Name), nil
	} else if volume.Secret != nil {
		return fmt.Sprintf("secret/%s/%s", namespace, volume.Secret.SecretName), nil
	} else {
		return "", errors.New("volume format must be ConfigMap or Secret")
	}
}

func findPrometheusPodConfigs(ctx context.Context, client kubernetes.Interface, namespace string) ([]podConfig, error) {
	var configs []podConfig
	pods, err := findPrometheusPods(ctx, client, namespace)
	if err != nil {
		return nil, err
	}
	var configCache = make(map[string]bool)

	for _, pod := range pods {
		// Extract volume of Prometheus config.
		volume, key, err := findPrometheusConfigVolumeAndKey(ctx, client, &pod)
		if err != nil {
			return nil, err
		}

		// See if we processed it already; if so, don't process again.
		ck, err := configKey(pod.Namespace, volume)
		if err != nil {
			return nil, err
		}
		_, ok := configCache[ck]
		if ok {
			continue
		}
		configCache[ck] = true

		// Extract Prometheus config bytes from volume.
		b, err := extractPrometheusConfigFromVolume(ctx, client, volume, pod.Namespace, key)
		if err != nil {
			return nil, err
		}
		configs = append(configs, podConfig{
			pod: &pod,
			b:   b,
		})
	}

	return configs, nil
}

func findPrometheusPods(ctx context.Context, client kubernetes.Interface, namespace string) ([]corev1.Pod, error) {
	var proms []corev1.Pod
	for k, v := range prometheusPodLabels {
		pods, err := util.GetPodListByLabels(client, namespace, map[string]string{
			k: v,
		})
		if err != nil {
			return nil, err
		}
		proms = append(proms, pods.Items...)
	}

	// If we still haven't found any Prometheus pods, make a last-ditch effort to
	// scrape the namespace for "prometheus" containers.
	if len(proms) == 0 {
		pods, err := client.CoreV1().Pods(namespace).List(ctx, v1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, pod := range pods.Items {
			for _, c := range pod.Spec.Containers {
				if c.Name == prometheusContainerName {
					proms = append(proms, pod)
				}
			}
		}
	}
	return proms, nil
}

func findPrometheusConfigPath(ctx context.Context, client kubernetes.Interface, pod *corev1.Pod) (string, error) {
	var path string
	var err error
	for _, container := range pod.Spec.Containers {
		for _, arg := range container.Args {
			// Prefer the config-reloader container config file as it normally
			// references the ConfigMap or Secret volume mount.
			// Fallback to the prometheus container if that's not found.
			if strings.HasPrefix(arg, prometheusConfigFlag) {
				path = strings.TrimPrefix(arg, prometheusConfigFlag)
			}
			if strings.HasPrefix(arg, configReloaderConfigFlag) {
				path = strings.TrimPrefix(arg, configReloaderConfigFlag)
			}
		}
		if container.Name == configReloaderContainerName {
			return path, nil
		}
	}
	if path == "" {
		err = fmt.Errorf("prometheus config path not found in pod: %s", pod.Name)
	}
	return path, err
}

func findPrometheusConfigVolumeAndKey(ctx context.Context, client kubernetes.Interface, pod *corev1.Pod) (*corev1.Volume, string, error) {
	path, err := findPrometheusConfigPath(ctx, client, pod)
	if err != nil {
		return nil, "", err
	}

	// Find the volumeMount the config path is pointing to.
	var volumeName = ""
	for _, container := range pod.Spec.Containers {
		for _, vm := range container.VolumeMounts {
			if strings.HasPrefix(path, vm.MountPath) {
				volumeName = vm.Name
				break
			}
		}
	}

	// Get the actual Volume from the name.
	for _, volume := range pod.Spec.Volumes {
		if volume.Name == volumeName {
			return &volume, filepath.Base(path), nil
		}
	}

	return nil, "", errors.New("volume for Prometheus config not found")
}

func extractPrometheusConfigFromVolume(ctx context.Context, client kubernetes.Interface, volume *corev1.Volume, namespace, key string) ([]byte, error) {
	var b []byte
	var ok bool
	// Check for Secret volume.
	if vs := volume.Secret; vs != nil {
		s, err := client.CoreV1().Secrets(namespace).Get(ctx, vs.SecretName, v1.GetOptions{})
		if err != nil {
			return nil, err
		}
		b, ok = s.Data[key]
		if !ok {
			return nil, fmt.Errorf("unable to find file key in secret: %s", key)
		}
	}
	// Check for ConfigMap volume.
	if vcm := volume.ConfigMap; vcm != nil {
		cm, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, vcm.Name, v1.GetOptions{})
		if err != nil {
			return nil, err
		}
		s, ok := cm.Data[key]
		b = []byte(s)
		if !ok {
			return nil, fmt.Errorf("unable to find file key in configmap: %s", key)
		}
	}
	return b, nil
}

func unmarshalPromConfigBytes(b []byte) (*promconfig.Config, error) {
	var config promconfig.Config
	// Unmarshal the data into a Prometheus config.
	if err := yaml.Unmarshal(b, &config); err == nil {
		return &config, nil
		// If there were errors, try gunziping the data.
	} else if content := http.DetectContentType(b); content == "application/x-gzip" {
		r, err := gzip.NewReader(bytes.NewBuffer(b))
		if err != nil {
			return &config, err
		}
		gunzipBytes, err := io.ReadAll(r)
		if err != nil {
			return &config, err
		}
		err = yaml.Unmarshal(gunzipBytes, &config)
		if err != nil {
			return nil, err
		}
		return &config, nil
	} else {
		return &config, err
	}
}
