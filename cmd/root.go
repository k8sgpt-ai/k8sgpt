package cmd

import (
	"os"
	"path/filepath"

	"github.com/k8sgpt-ai/k8sgpt/cmd/generate"
	"k8s.io/client-go/util/homedir"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/cmd/analyze"
	"github.com/k8sgpt-ai/k8sgpt/cmd/auth"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	kubecontext string
	kubeconfig  string
	version     string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k8sgpt",
	Short: "Kubernetes debugging powered by AI",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(v string) {
	version = v
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	var kubeconfigPath string
	if home := homedir.HomeDir(); home != "" {
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}
	rootCmd.AddCommand(auth.AuthCmd)
	rootCmd.AddCommand(analyze.AnalyzeCmd)
	rootCmd.AddCommand(generate.GenerateCmd)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.k8sgpt.yaml)")
	rootCmd.PersistentFlags().StringVar(&kubecontext, "kubecontext", "", "Kubernetes context to use. Only required if out-of-cluster.")
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", kubeconfigPath, "Path to a kubeconfig. Only required if out-of-cluster.")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	//Initialise the kubeconfig
	kubernetesClient, err := kubernetes.NewClient(kubecontext, kubeconfig)
	if err != nil {
		color.Red("Error initialising kubernetes client: %v", err)
		os.Exit(1)
	}

	viper.Set("kubernetesClient", kubernetesClient)

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".k8sgpt.git" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".k8sgpt")

		viper.SafeWriteConfig()
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		//	fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
