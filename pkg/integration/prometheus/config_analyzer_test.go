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

package prometheus

import (
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// gzipBytes returns b compressed with gzip so that http.DetectContentType
// classifies the result as "application/x-gzip".
func gzipBytes(t *testing.T, b []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(b); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buf.Bytes()
}

// TestUnmarshalPromConfigBytesGzipInvalid guards the contract that the helper
// never returns a nil *Config. A payload that is detected as gzip but gunzips
// to invalid YAML previously returned (nil, err), which made both analyzers
// panic when they dereferenced config.ScrapeConfigs.
func TestUnmarshalPromConfigBytesGzipInvalid(t *testing.T) {
	// Invalid YAML for a Prometheus config (a bare scalar, not a mapping).
	b := gzipBytes(t, []byte("::: not valid yaml :::"))

	require.Equal(t, "application/x-gzip", http.DetectContentType(b),
		"fixture must be detected as gzip to exercise the gunzip path")

	config, err := unmarshalPromConfigBytes(b)
	require.Error(t, err)
	require.NotNil(t, config, "helper must never return a nil config")
	// Reading ScrapeConfigs must be safe (this is the previously panicking deref).
	require.Empty(t, config.ScrapeConfigs)
}

// TestPrometheusAnalyzersGzipInvalidConfig drives both Prometheus analyzers
// end to end against a pod whose config volume holds a gzip-detected but
// invalid Prometheus configuration. They must surface failures instead of
// panicking.
func TestPrometheusAnalyzersGzipInvalidConfig(t *testing.T) {
	const (
		namespace  = "monitoring"
		configKey  = "prometheus.yaml"
		volumeName = "config"
		mountPath  = "/etc/prometheus"
		configMap  = "prometheus-config"
	)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "prometheus-0",
			Namespace: namespace,
			Labels:    map[string]string{"app": "prometheus"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: configReloaderContainerName,
					Args: []string{configReloaderConfigFlag + mountPath + "/" + configKey},
					VolumeMounts: []corev1.VolumeMount{
						{Name: volumeName, MountPath: mountPath},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: volumeName,
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: configMap},
						},
					},
				},
			},
		},
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMap,
			Namespace: namespace,
		},
		// http.DetectContentType only inspects bytes, so gzip-detected but
		// invalid YAML reaches the gunzip branch of unmarshalPromConfigBytes.
		Data: map[string]string{
			configKey: string(gzipBytes(t, []byte("::: not valid yaml :::"))),
		},
	}

	newAnalyzer := func() common.Analyzer {
		return common.Analyzer{
			Client:    &kubernetes.Client{Client: fake.NewSimpleClientset(pod, cm)},
			Context:   context.Background(),
			Namespace: namespace,
		}
	}

	t.Run("ConfigAnalyzer", func(t *testing.T) {
		results, err := (&ConfigAnalyzer{}).Analyze(newAnalyzer())
		require.NoError(t, err)
		require.NotEmpty(t, results, "expected a validation failure for the invalid config")
	})

	t.Run("RelabelAnalyzer", func(t *testing.T) {
		results, err := (&RelabelAnalyzer{}).Analyze(newAnalyzer())
		require.NoError(t, err)
		// No relabel report is expected; the point is that it does not panic.
		require.Empty(t, results)
	})
}
