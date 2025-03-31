package keda

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/kedacore/keda/v2/pkg/generated/clientset/versioned/typed/keda/v1alpha1"
	helmclient "github.com/mittwald/go-helm-client"
	"github.com/spf13/viper"
	"helm.sh/helm/v3/pkg/repo"
)

var (
	Repo          = getEnv("KEDA_REPO", "https://kedacore.github.io/charts")
	Version       = getEnv("KEDA_VERSION", "2.11.2")
	ChartName     = getEnv("KEDA_CHART_NAME", "keda")
	RepoShortName = getEnv("KEDA_REPO_SHORT_NAME", "keda")
	ReleaseName   = getEnv("KEDA_RELEASE_NAME", "keda-k8sgpt")
)

type Keda struct {
	helm helmclient.Client
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func NewKeda() *Keda {
	helmClient, err := helmclient.New(&helmclient.Options{})
	if err != nil {
		panic(err)
	}
	return &Keda{
		helm: helmClient,
	}
}

func (k *Keda) Deploy(namespace string) error {
	// Add the repository
	chartRepo := repo.Entry{
		Name: RepoShortName,
		URL:  Repo,
	}
	// Add a chart-repository to the client.
	if err := k.helm.AddOrUpdateChartRepo(chartRepo); err != nil {
		panic(err)
	}

	chartSpec := helmclient.ChartSpec{
		ReleaseName: ReleaseName,
		ChartName:   fmt.Sprintf("%s/%s", RepoShortName, ChartName),
		Namespace:   namespace,

		//TODO: All of this should be configurable
		UpgradeCRDs:     true,
		Wait:            false,
		Timeout:         300,
		CreateNamespace: true,
	}

	// Install a chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	if _, err := k.helm.InstallOrUpgradeChart(context.Background(), &chartSpec, nil); err != nil {
		return err
	}

	return nil
}

func (k *Keda) UnDeploy(namespace string) error {
	kubecontext := viper.GetString("kubecontext")
	kubeconfig := viper.GetString("kubeconfig")
	client, err := kubernetes.NewClient(kubecontext, kubeconfig)
	if err != nil {
		// TODO: better error handling
		color.Red("Error initialising kubernetes client: %v", err)
		os.Exit(1)
	}

	kedaNamespace, _ := k.GetNamespace()
	color.Blue(fmt.Sprintf("Keda namespace: %s\n", kedaNamespace))

	kClient, _ := v1alpha1.NewForConfig(client.Config)

	scaledObjectList, _ := kClient.ScaledObjects("").List(context.Background(), v1.ListOptions{})
	scaledJobList, _ := kClient.ScaledJobs("").List(context.Background(), v1.ListOptions{})
	triggerAuthenticationList, _ := kClient.TriggerAuthentications("").List(context.Background(), v1.ListOptions{})
	clusterTriggerAuthenticationsList, _ := kClient.ClusterTriggerAuthentications().List(context.Background(), v1.ListOptions{})

	// Before uninstalling the Helm chart, we need to delete Keda resources
	for _, scaledObject := range scaledObjectList.Items {
		err := kClient.ScaledObjects(scaledObject.Namespace).Delete(context.Background(), scaledObject.Name, v1.DeleteOptions{})
		if err != nil {
			fmt.Printf("Error deleting scaledObject %s: %v\n", scaledObject.Name, err)
		} else {
			fmt.Printf("Deleted scaledObject %s in namespace %s\n", scaledObject.Name, scaledObject.Namespace)
		}
	}

	for _, scaledJob := range scaledJobList.Items {
		err := kClient.ScaledJobs(scaledJob.Namespace).Delete(context.Background(), scaledJob.Name, v1.DeleteOptions{})
		if err != nil {
			fmt.Printf("Error deleting scaledJob %s: %v\n", scaledJob.Name, err)
		} else {
			fmt.Printf("Deleted scaledJob %s in namespace %s\n", scaledJob.Name, scaledJob.Namespace)
		}
	}

	for _, triggerAuthentication := range triggerAuthenticationList.Items {
		err := kClient.TriggerAuthentications(triggerAuthentication.Namespace).Delete(context.Background(), triggerAuthentication.Name, v1.DeleteOptions{})
		if err != nil {
			fmt.Printf("Error deleting triggerAuthentication %s: %v\n", triggerAuthentication.Name, err)
		} else {
			fmt.Printf("Deleted triggerAuthentication %s in namespace %s\n", triggerAuthentication.Name, triggerAuthentication.Namespace)
		}
	}

	for _, clusterTriggerAuthentication := range clusterTriggerAuthenticationsList.Items {
		err := kClient.ClusterTriggerAuthentications().Delete(context.Background(), clusterTriggerAuthentication.Name, v1.DeleteOptions{})
		if err != nil {
			fmt.Printf("Error deleting clusterTriggerAuthentication %s: %v\n", clusterTriggerAuthentication.Name, err)
		} else {
			fmt.Printf("Deleted clusterTriggerAuthentication %s\n", clusterTriggerAuthentication.Name)
		}
	}

	chartSpec := helmclient.ChartSpec{
		ReleaseName: ReleaseName,
		ChartName:   fmt.Sprintf("%s/%s", RepoShortName, ChartName),
		Namespace:   namespace,
		UpgradeCRDs: true,
		Wait:        false,
		Timeout:     300,
	}
	// Uninstall the chart release.
	// Note that helmclient.Options.Namespace should ideally match the namespace in chartSpec.Namespace.
	if err := k.helm.UninstallRelease(&chartSpec); err != nil {
		return err
	}
	return nil
}

func (k *Keda) AddAnalyzer(mergedMap *map[string]common.IAnalyzer) {
	(*mergedMap)["ScaledObject"] = &ScaledObjectAnalyzer{}
}

func (k *Keda) GetAnalyzerName() []string {
	return []string{
		"ScaledObject",
	}
}

func (k *Keda) GetNamespace() (string, error) {
	releases, err := k.helm.ListDeployedReleases()
	if err != nil {
		return "", err
	}
	for _, rel := range releases {
		if rel.Name == ReleaseName {
			return rel.Namespace, nil
		}
	}
	return "", status.Error(codes.NotFound, "keda release not found")
}

func (k *Keda) OwnsAnalyzer(analyzer string) bool {
	for _, a := range k.GetAnalyzerName() {
		if analyzer == a {
			return true
		}
	}
	return false
}

func (k *Keda) isFilterActive() bool {
	activeFilters := viper.GetStringSlice("active_filters")

	for _, filter := range k.GetAnalyzerName() {
		for _, af := range activeFilters {
			if af == filter {
				return true
			}
		}
	}

	return false
}

func (k *Keda) isDeployed() bool {
	kubecontext := viper.GetString("kubecontext")
	kubeconfig := viper.GetString("kubeconfig")
	client, err := kubernetes.NewClient(kubecontext, kubeconfig)
	if err != nil {
		// TODO: better error handling
		color.Red("Error initialising kubernetes client: %v", err)
		os.Exit(1)
	}
	groups, _, err := client.Client.Discovery().ServerGroupsAndResources()
	if err != nil {
		// TODO: better error handling
		color.Red("Error initialising discovery client: %v", err)
		os.Exit(1)
	}

	for _, group := range groups {
		if group.Name == "keda.sh" {
			return true
		}
	}

	return false
}

func (k *Keda) IsActivate() bool {
	return k.isFilterActive() && k.isDeployed()
}
