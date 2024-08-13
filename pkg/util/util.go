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

package util

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k "k8s.io/client-go/kubernetes"
)

var anonymizePattern = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;':\",./<>?")

func GetParent(client *kubernetes.Client, meta metav1.ObjectMeta) (string, bool) {
	if meta.OwnerReferences != nil {
		for _, owner := range meta.OwnerReferences {
			switch owner.Kind {
			case "ReplicaSet":
				rs, err := client.GetClient().AppsV1().ReplicaSets(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if rs.OwnerReferences != nil {
					return GetParent(client, rs.ObjectMeta)
				}
				return "ReplicaSet/" + rs.Name, true

			case "Deployment":
				dep, err := client.GetClient().AppsV1().Deployments(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if dep.OwnerReferences != nil {
					return GetParent(client, dep.ObjectMeta)
				}
				return "Deployment/" + dep.Name, true

			case "StatefulSet":
				sts, err := client.GetClient().AppsV1().StatefulSets(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if sts.OwnerReferences != nil {
					return GetParent(client, sts.ObjectMeta)
				}
				return "StatefulSet/" + sts.Name, true

			case "DaemonSet":
				ds, err := client.GetClient().AppsV1().DaemonSets(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if ds.OwnerReferences != nil {
					return GetParent(client, ds.ObjectMeta)
				}
				return "DaemonSet/" + ds.Name, true

			case "Ingress":
				ds, err := client.GetClient().NetworkingV1().Ingresses(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if ds.OwnerReferences != nil {
					return GetParent(client, ds.ObjectMeta)
				}
				return "Ingress/" + ds.Name, true

			case "MutatingWebhookConfiguration":
				mw, err := client.GetClient().AdmissionregistrationV1().MutatingWebhookConfigurations().Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if mw.OwnerReferences != nil {
					return GetParent(client, mw.ObjectMeta)
				}
				return "MutatingWebhook/" + mw.Name, true

			case "ValidatingWebhookConfiguration":
				vw, err := client.GetClient().AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if vw.OwnerReferences != nil {
					return GetParent(client, vw.ObjectMeta)
				}
				return "ValidatingWebhook/" + vw.Name, true
			}
		}
	}
	return "", false
}

func RemoveDuplicates(slice []string) ([]string, []string) {
	set := make(map[string]bool)
	duplicates := []string{}
	for _, val := range slice {
		if _, ok := set[val]; !ok {
			set[val] = true
		} else {
			duplicates = append(duplicates, val)
		}
	}
	uniqueSlice := make([]string, 0, len(set))
	for val := range set {
		uniqueSlice = append(uniqueSlice, val)
	}
	return uniqueSlice, duplicates
}

func SliceDiff(source, dest []string) []string {
	mb := make(map[string]struct{}, len(dest))
	for _, x := range dest {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range source {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func MaskString(input string) string {
	key := make([]byte, len(input))
	result := make([]rune, len(input))
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	for i := range result {
		result[i] = anonymizePattern[int(key[i])%len(anonymizePattern)]
	}
	return base64.StdEncoding.EncodeToString([]byte(string(result)))
}

func ReplaceIfMatch(text string, pattern string, replacement string) string {
	re := regexp.MustCompile(fmt.Sprintf(`%s(\b)`, pattern))
	if re.MatchString(text) {
		text = re.ReplaceAllString(text, replacement)
	}
	return text
}

func GetCacheKey(provider string, language string, sEnc string) string {
	data := fmt.Sprintf("%s-%s-%s", provider, language, sEnc)

	hash := sha256.Sum256([]byte(data))

	return hex.EncodeToString(hash[:])
}

func GetPodListByLabels(client k.Interface,
	namespace string,
	labels map[string]string,
) (*v1.PodList, error) {
	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: labels,
		}),
	})
	if err != nil {
		return nil, err
	}

	return pods, nil
}

func FileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func EnsureDirExists(dir string) error {
	err := os.MkdirAll(dir, 0o755)

	if errors.Is(err, os.ErrExist) {
		return nil
	}

	return err
}

func MapToString(m map[string]string) string {
	// Handle empty map case
	if len(m) == 0 {
		return ""
	}

	var pairs []string
	for k, v := range m {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}

	// Efficient string joining
	return strings.Join(pairs, ",")
}

func LabelsIncludeAny(predefinedSelector, Labels map[string]string) bool {
	// Check if any label in the predefinedSelector exists in Labels
	for key := range predefinedSelector {
		if _, exists := Labels[key]; exists {
			return true
		}
	}

	return false
}

func FetchLatestEvent(ctx context.Context, kubernetesClient *kubernetes.Client, namespace string, name string) (*v1.Event, error) {

	// get the list of events
	events, err := kubernetesClient.GetClient().CoreV1().Events(namespace).List(ctx,
		metav1.ListOptions{
			FieldSelector: "involvedObject.name=" + name,
		})

	if err != nil {
		return nil, err
	}
	// find most recent event
	var latestEvent *v1.Event
	for _, event := range events.Items {
		if latestEvent == nil {
			// this is required, as a pointer to a loop variable would always yield the latest value in the range
			e := event
			latestEvent = &e
		}
		if event.LastTimestamp.After(latestEvent.LastTimestamp.Time) {
			// this is required, as a pointer to a loop variable would always yield the latest value in the range
			e := event
			latestEvent = &e
		}
	}
	return latestEvent, nil
}

// NewHeaders parses a slice of strings in the format "key:value" into []http.Header
// It handles headers with the same key by appending values
func NewHeaders(customHeaders []string) []http.Header {
	headers := make(map[string][]string)

	for _, header := range customHeaders {
		vals := strings.SplitN(header, ":", 2)
		if len(vals) != 2 {
			//TODO: Handle error instead of ignoring it
			continue
		}
		key := strings.TrimSpace(vals[0])
		value := strings.TrimSpace(vals[1])

		if _, ok := headers[key]; !ok {
			headers[key] = []string{}
		}
		headers[key] = append(headers[key], value)
	}

	// Convert map to []http.Header format
	var result []http.Header
	for key, values := range headers {
		header := make(http.Header)
		for _, value := range values {
			header.Add(key, value)
		}
		result = append(result, header)
	}

	return result
}

func LabelStrToSelector(labelStr string) labels.Selector {
	if labelStr == "" {
		return nil
	}
	labelSelectorMap := make(map[string]string)
	for _, s := range strings.Split(labelStr, ",") {
		parts := strings.SplitN(s, "=", 2)
		if len(parts) == 2 {
			labelSelectorMap[parts[0]] = parts[1]
		}
	}
	return labels.SelectorFromSet(labels.Set(labelSelectorMap))
}
