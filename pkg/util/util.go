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
	"os"
	"regexp"
	"strings"

	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k "k8s.io/client-go/kubernetes"
)

var anonymizePattern = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;':\",./<>?")

func SliceContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

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
				return "ReplicaSet/" + rs.Name, false

			case "Deployment":
				dep, err := client.GetClient().AppsV1().Deployments(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if dep.OwnerReferences != nil {
					return GetParent(client, dep.ObjectMeta)
				}
				return "Deployment/" + dep.Name, false

			case "StatefulSet":
				sts, err := client.GetClient().AppsV1().StatefulSets(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if sts.OwnerReferences != nil {
					return GetParent(client, sts.ObjectMeta)
				}
				return "StatefulSet/" + sts.Name, false

			case "DaemonSet":
				ds, err := client.GetClient().AppsV1().DaemonSets(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if ds.OwnerReferences != nil {
					return GetParent(client, ds.ObjectMeta)
				}
				return "DaemonSet/" + ds.Name, false

			case "Ingress":
				ds, err := client.GetClient().NetworkingV1().Ingresses(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if ds.OwnerReferences != nil {
					return GetParent(client, ds.ObjectMeta)
				}
				return "Ingress/" + ds.Name, false

			case "MutatingWebhookConfiguration":
				mw, err := client.GetClient().AdmissionregistrationV1().MutatingWebhookConfigurations().Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if mw.OwnerReferences != nil {
					return GetParent(client, mw.ObjectMeta)
				}
				return "MutatingWebhook/" + mw.Name, false

			case "ValidatingWebhookConfiguration":
				vw, err := client.GetClient().AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if vw.OwnerReferences != nil {
					return GetParent(client, vw.ObjectMeta)
				}
				return "ValidatingWebhook/" + vw.Name, false
			}
		}
	}
	return meta.Name, false
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
	labels map[string]string) (*v1.PodList, error) {
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
	err := os.MkdirAll(dir, 0755)

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
