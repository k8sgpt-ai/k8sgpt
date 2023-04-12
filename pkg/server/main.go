package server

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analysis"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/spf13/viper"
	"net/http"
	"os"
)

type K8sGPTServer struct {
	Port    string
	Backend string
	Key     string
	Token   string
	Output  string
}

type Result struct {
	Analysis []analysis.Analysis `json:"analysis"`
}

func (s *K8sGPTServer) analyzeHandler(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	explain := getBoolParam(r.URL.Query().Get("explain"))
	anonymize := getBoolParam(r.URL.Query().Get("anonymize"))
	nocache := getBoolParam(r.URL.Query().Get("nocache"))
	language := r.URL.Query().Get("language")

	// get ai configuration
	var configAI ai.AIConfiguration
	err := viper.UnmarshalKey("ai", &configAI)
	if err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}

	if len(configAI.Providers) == 0 {
		color.Red("Error: AI provider not specified in configuration. Please run k8sgpt auth")
		os.Exit(1)
	}

	var aiProvider ai.AIProvider
	for _, provider := range configAI.Providers {
		if s.Backend == provider.Name {
			aiProvider = provider
			break
		}
	}

	if aiProvider.Name == "" {
		color.Red("Error: AI provider %s not specified in configuration. Please run k8sgpt auth", s.Backend)
		os.Exit(1)
	}

	aiClient := ai.NewClient(aiProvider.Name)
	if err := aiClient.Configure(aiProvider.Password, aiProvider.Model, language); err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}

	ctx := context.Background()
	// Get kubernetes client from viper

	kubecontext := viper.GetString("kubecontext")
	kubeconfig := viper.GetString("kubeconfig")
	client, err := kubernetes.NewClient(kubecontext, kubeconfig)
	if err != nil {
		color.Red("Error initialising kubernetes client: %v", err)
		os.Exit(1)
	}

	config := &analysis.Analysis{
		Namespace: namespace,
		Explain:   explain,
		AIClient:  aiClient,
		Client:    client,
		Context:   ctx,
		NoCache:   nocache,
	}

	err = config.RunAnalysis()
	if err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}

	if explain {
		err := config.GetAIResults(s.Output, anonymize)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}
	}

	output, err := config.JsonOutput()
	if err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}
	fmt.Fprintf(w, string(output))

}

func (s *K8sGPTServer) Serve() error {
	http.HandleFunc("/analyze", s.analyzeHandler)
	err := http.ListenAndServe(":"+s.Port, nil)
	if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		return err
	}
	return nil
}

func getBoolParam(param string) bool {
	if param == "true" {
		return true
	}
	return false
}
