package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analyzer"
	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"strings"
)

type K8sGPTServer struct {
	Port    string
	Backend string
	Key     string
	Token   string
}

type Result struct {
	Analysis []analyzer.Analysis `json:"analysis"`
}

func (s *K8sGPTServer) analyzeHandler(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	ex := r.URL.Query().Get("explain")

	explain := false

	if ex == "true" {
		explain = true
	}

	output := Result{}

	var aiClient ai.IAI
	switch s.Backend {
	case "openai":
		aiClient = &ai.OpenAIClient{}
		if err := aiClient.Configure(s.Token, "english"); err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}
	default:
		color.Red("Backend not supported")
		os.Exit(1)
	}

	ctx := context.Background()
	// Get kubernetes client from viper
	client := viper.Get("kubernetesClient").(*kubernetes.Client)
	// Analysis configuration
	config := &analyzer.AnalysisConfiguration{
		Namespace: namespace,
		Explain:   explain,
	}

	var analysisResults *[]analyzer.Analysis = &[]analyzer.Analysis{}
	if err := analyzer.RunAnalysis(ctx, []string{}, config, client,
		aiClient, analysisResults); err != nil {
		color.Red("Error: %v", err)
	}

	fmt.Println(analysisResults)
	if len(*analysisResults) == 0 {
		fmt.Fprintf(w, "{ \"status\": \"OK\" }")
	}

	var bar = progressbar.Default(int64(len(*analysisResults)))
	if !explain {
		bar.Clear()
	}
	var printOutput []analyzer.Analysis

	for _, analysis := range *analysisResults {

		if explain {
			parsedText, err := analyzer.ParseViaAI(ctx, config, aiClient, analysis.Error)
			if err != nil {
				// Check for exhaustion
				if strings.Contains(err.Error(), "status code: 429") {
					fmt.Fprintf(w, "Exhausted API quota. Please try again later")
					os.Exit(1)
				}
				color.Red("Error: %v", err)
				continue
			}
			analysis.Details = parsedText
			bar.Add(1)
		}
		printOutput = append(printOutput, analysis)

		analysis.Error = analysis.Error[0:]
		output.Analysis = append(output.Analysis, analysis)
	}
	j, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}
	fmt.Fprintf(w, "%s", j)

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
