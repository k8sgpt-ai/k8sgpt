package config

import "github.com/spf13/viper"

type KubernetesSettings struct {
	Context string
	Config  string
}

func LoadKubernetesSettings() KubernetesSettings {
	kubecontext := viper.GetString("kubecontext")
	kubeconfig := viper.GetString("kubeconfig")

	return KubernetesSettings{
		Context: kubecontext,
		Config:  kubeconfig,
	}
}

func SetKubernetesSettings(kubernetesSettings KubernetesSettings) {
	viper.Set("kubecontext", kubernetesSettings.Context)
	viper.Set("kubeconfig", kubernetesSettings.Config)
}
