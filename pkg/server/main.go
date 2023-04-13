package server

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analysis"
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

	config, err := analysis.NewAnalysis(s.Backend, language, []string{}, namespace, nocache, explain)
	if err != nil {
		fmt.Fprintf(w, err.Error())
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
	color.Green("Starting server on port " + s.Port)
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
