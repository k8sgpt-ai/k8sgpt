package local

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	clientGoScheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	fakerest "k8s.io/client-go/rest/fake"
	"k8s.io/client-go/testing"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	fakeCtrlclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"strings"
)

var sch *runtime.Scheme

type fakerAttributes struct {
	Resource string
	Obj      runtime.Object
	GVK      schema.GroupVersionKind
	Lister   runtime.Object
}

var (
	fakers = []fakerAttributes{
		{
			Resource: "pods",
			Obj:      &corev1.Pod{},
			Lister:   &corev1.PodList{},
			GVK:      schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
		},
		{
			Resource: "services",
			Obj:      &corev1.Service{},
			Lister:   &corev1.ServiceList{},
			GVK:      schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
		},
		{
			Resource: "configmaps",
			Obj:      &corev1.ConfigMap{},
			Lister:   &corev1.ConfigMapList{},
			GVK:      schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"},
		},
		{
			Resource: "secrets",
			Obj:      &corev1.Secret{},
			Lister:   &corev1.SecretList{},
			GVK:      schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"},
		},
		{
			Resource: "endpoints",
			Obj:      &corev1.Endpoints{},
			Lister:   &corev1.EndpointsList{},
			GVK:      schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Endpoints"},
		},
		{
			Resource: "events",
			Obj:      &corev1.Event{},
			Lister:   &corev1.EventList{},
			GVK:      schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Event"},
		},
		{
			Resource: "namespaces",
			Obj:      &corev1.Namespace{},
			Lister:   &corev1.NamespaceList{},
			GVK:      schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Namespace"},
		},
		{
			Resource: "nodes",
			Obj:      &corev1.Node{},
			Lister:   &corev1.NodeList{},
			GVK:      schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Node"},
		},
		{
			Resource: "persistentvolumes",
			Obj:      &corev1.PersistentVolume{},
			Lister:   &corev1.PersistentVolumeList{},
			GVK:      schema.GroupVersionKind{Group: "", Version: "v1", Kind: "PersistentVolume"},
		},
		{
			Resource: "persistentvolumeclaims",
			Obj:      &corev1.PersistentVolumeClaim{},
			Lister:   &corev1.PersistentVolumeClaimList{},
			GVK:      schema.GroupVersionKind{Group: "", Version: "v1", Kind: "PersistentVolumeClaim"},
		},
		{
			Resource: "replicationcontrollers",
			Obj:      &corev1.ReplicationController{},
			Lister:   &corev1.ReplicationControllerList{},
			GVK:      schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ReplicationController"},
		},
		{
			Resource: "serviceaccounts",
			Obj:      &corev1.ServiceAccount{},
			Lister:   &corev1.ServiceAccountList{},
			GVK:      schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"},
		},
		{
			Resource: "deployments",
			Obj:      &appsv1.Deployment{},
			Lister:   &appsv1.DeploymentList{},
			GVK:      schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		},
		{
			Resource: "daemonsets",
			Obj:      &appsv1.DaemonSet{},
			Lister:   &appsv1.DaemonSetList{},
			GVK:      schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"},
		},
		{
			Resource: "replicasets",
			Obj:      &appsv1.ReplicaSet{},
			Lister:   &appsv1.ReplicaSetList{},
			GVK:      schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "ReplicaSet"},
		},
		{
			Resource: "statefulsets",
			Obj:      &appsv1.StatefulSet{},
			Lister:   &appsv1.StatefulSetList{},
			GVK:      schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
		},
		{
			Resource: "ingresses",
			Obj:      &networkingv1.Ingress{},
			Lister:   &networkingv1.IngressList{},
			GVK:      schema.GroupVersionKind{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress"},
		},
		{
			Resource: "networkpolicies",
			Obj:      &networkingv1.NetworkPolicy{},
			Lister:   &networkingv1.NetworkPolicyList{},
			GVK:      schema.GroupVersionKind{Group: "networking.k8s.io", Version: "v1", Kind: "NetworkPolicy"},
		},
		{
			Resource: "ingressclasses",
			Obj:      &networkingv1.IngressClass{},
			Lister:   &networkingv1.IngressClassList{},
			GVK:      schema.GroupVersionKind{Group: "networking.k8s.io", Version: "v1", Kind: "IngressClass"},
		},
		{
			Resource: "mutaingwebhookconfigurations",
			Obj:      &admissionregistrationv1.MutatingWebhookConfiguration{},
			Lister:   &admissionregistrationv1.MutatingWebhookConfigurationList{},
			GVK:      schema.GroupVersionKind{Group: "admissionregistration.k8s.io", Version: "v1", Kind: "MutatingWebhookConfiguration"},
		},
		{
			Resource: "validatingwebhookconfigurations",
			Obj:      &admissionregistrationv1.ValidatingWebhookConfiguration{},
			Lister:   &admissionregistrationv1.ValidatingWebhookConfigurationList{},
			GVK:      schema.GroupVersionKind{Group: "admissionregistration.k8s.io", Version: "v1", Kind: "ValidatingWebhookConfiguration"},
		},
		{
			Resource: "poddisruptionbudgets",
			Obj:      &policyv1.PodDisruptionBudget{},
			Lister:   &policyv1.PodDisruptionBudgetList{},
			GVK:      schema.GroupVersionKind{Group: "policy", Version: "v1", Kind: "PodDisruptionBudget"},
		},
	}
)

func findLogFiles(rcaPath string, namespace string, name string) []string {
	globPattern := fmt.Sprintf("kubectl_logs_--namespace_%s_%s.log", namespace, name)
	files, err := filepath.Glob(rcaPath + "/" + globPattern)
	if err != nil {
		return []string{}
	}
	return files
}

func findFilesForResource(rcaPath string, resource string, namespace string) []string {
	resource = strings.ToLower(resource)
	namespace = strings.ToLower(namespace)
	globPattern := fmt.Sprintf("kubectl_get_%s_*-o_yaml.log", resource)
	if namespace != "" {
		globPattern = fmt.Sprintf("kubectl_get_%s_--namespace_%s_*-o_yaml.log", resource, namespace)
	}

	files, err := filepath.Glob(rcaPath + "/" + globPattern)
	if err != nil {
		klog.ErrorS(err, "failed to determine the files container information for the resource", "resource", resource, "namespace", namespace, "pattern", globPattern)
		return []string{}
	}
	return files
}

func getResourceFromFile[T runtime.Object](file string, objType T) error {
	data, err := os.Open(file)
	if err != nil {
		return err
	}
	defer func() {
		if err := data.Close(); err != nil {
			klog.ErrorS(err, "failed to close file", "file", file)
		}
	}()

	scanner := bufio.NewScanner(data)
	var buffer []byte
	buf := bytes.NewBuffer(buffer)
	writer := bufio.NewWriter(buf)
	for scanner.Scan() {
		l := scanner.Bytes()
		if string(l) == "---------------------------------------------------------------------" {
			break
		}
		if _, err := writer.WriteString(fmt.Sprintf("%s\n", string(l))); err != nil {
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		return err
	}

	decoder := yaml.NewYAMLOrJSONDecoder(bufio.NewReader(bytes.NewBuffer(buf.Bytes())), 100)
	if err := decoder.Decode(objType); err != nil {
		return err
	}
	return nil
}

func dataFetcher[T runtime.Object](objType T, gvk schema.GroupVersionKind, resourceKind string, rcaPath string, action testing.Action) (bool, []T, error) {
	files := findFilesForResource(rcaPath, resourceKind, action.GetNamespace())
	list := &unstructured.UnstructuredList{
		Items: make([]unstructured.Unstructured, 0),
	}
	var items []T
	if len(files) == 0 && action.GetVerb() == "get" {
		klog.InfoS("Returning from fake client", "status", http.StatusNotFound, "files", files, "action", action)
		return true, nil, errors.NewNotFound(action.GetResource().GroupResource(), action.(testing.GetAction).GetName())
	}
	if len(files) == 0 {
		return true, nil, nil
	}
	for _, file := range files {
		err := getResourceFromFile(file, list)
		if err != nil {
			return true, nil, err
		}
		for _, d := range list.Items {
			obj := reflect.New(reflect.Indirect(reflect.ValueOf(objType)).Type())
			d.SetGroupVersionKind(gvk)
			if err := sch.Convert(&d, obj.Interface(), nil); err != nil {
				return true, nil, err
			}
			if action.GetVerb() == "get" && d.GetName() == action.(testing.GetAction).GetName() && d.GetNamespace() == action.(testing.GetAction).GetNamespace() {
				d.SetGroupVersionKind(gvk)
				return true, []T{obj.Interface().(T)}, nil
			}
			items = append(items, obj.Interface().(T))
		}
	}
	return true, items, nil
}

func GetLocalClient(rcaPath string) (kubernetes.Interface, ctrl.Client) {
	sch = runtime.NewScheme()
	_ = scheme.AddToScheme(sch)
	_ = apiextensionsv1.AddToScheme(sch)
	_ = clientGoScheme.AddToScheme(sch)

	fakeClient := fake.NewSimpleClientset()

	for _, attr := range fakers {
		for _, verb := range []string{"get", "list"} {
			fakeClient.PrependReactor(verb, attr.Resource, func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				if action.GetSubresource() == "log" {
					return false, nil, nil
				}
				handled, items, err := dataFetcher(attr.Obj, attr.GVK, attr.Resource, rcaPath, action)
				if err != nil {
					return handled, nil, err
				}
				if action.GetVerb() == "get" {
					return handled, items[0], nil
				}
				entries := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(attr.Obj).Elem()), 0, 0)
				for _, item := range items {
					entries = reflect.Append(entries, reflect.Indirect(reflect.ValueOf(item)))
				}
				reflect.Indirect(reflect.ValueOf(attr.Lister)).FieldByName("Items").Set(entries)
				return handled, attr.Lister, nil
			})
		}
	}

	fakeClient.PrependProxyReactor("*", func(action testing.Action) (handled bool, ret rest.ResponseWrapper, err error) {
		if action.GetSubresource() == "log" {
			apiPath := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/log", action.GetNamespace(), action.(testing.GetLogActionImpl).GetPodName())
			var attributes []string
			logAction := action.(testing.GetLogActionImpl)
			if logAction.ContainerName != "" {
				attributes = append(attributes, fmt.Sprintf("container=%s", logAction.ContainerName))
			}
			if logAction.SinceSeconds != nil {
				attributes = append(attributes, fmt.Sprintf("sinceSeconds=%ds", logAction.SinceSeconds))
			}

			if logAction.SinceTime != nil {
				attributes = append(attributes, fmt.Sprintf("sinceTime=%s", logAction.SinceTime.String()))
			}
			apiPath = fmt.Sprintf("%s?%s", apiPath, strings.Join(attributes, "&"))
			resp := &fakerest.RESTClient{
				Client: fakerest.CreateHTTPClient(func(request *http.Request) (*http.Response, error) {
					files := findLogFiles(rcaPath, action.GetNamespace(), action.(testing.GetActionImpl).GetName())
					if len(files) == 0 {
						resp := &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader("")),
						}
						return resp, nil
					}
					var buffer []byte
					buf := bytes.NewBuffer(buffer)
					writer := bufio.NewWriter(buf)
					for _, file := range files {
						data, err := os.ReadFile(file)
						if err != nil {
							return nil, err
						}
						_, _ = writer.Write(data)
						_, _ = writer.WriteString("\n")
					}
					_ = writer.Flush()
					resp := &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(buf.String())),
					}
					return resp, nil
				}),
				NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
				GroupVersion:         action.GetResource().GroupVersion(),
				VersionedAPIPath:     apiPath,
			}
			return true, resp.Request(), nil
		}
		return false, nil, nil
	})

	// Special handling for pod log extraction mechanism

	return fakeClient, fakeCtrlclient.NewClientBuilder().WithScheme(sch).WithRuntimeObjects().Build()
}
