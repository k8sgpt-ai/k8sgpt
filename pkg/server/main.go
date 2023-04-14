package server

import (
	json "encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/k8sgpt-ai/k8sgpt/pkg/analysis"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	Port    string
	Backend string
	Key     string
	Token   string
	Output  string
}

type Health struct {
	Status  string `json:"status"`
	Success int    `json:"success"`
	Failure int    `json:"failure"`
}

var health = Health{
	Status:  "ok",
	Success: 0,
	Failure: 0,
}

type Result struct {
	Analysis []analysis.Analysis `json:"analysis"`
}

func (s *Config) analyzeHandler(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	explain := getBoolParam(r.URL.Query().Get("explain"))
	anonymize := getBoolParam(r.URL.Query().Get("anonymize"))
	nocache := getBoolParam(r.URL.Query().Get("nocache"))
	language := r.URL.Query().Get("language")

	config, err := analysis.NewAnalysis(s.Backend, language, []string{}, namespace, nocache, explain)
	if err != nil {
		health.Failure++
		fmt.Fprintf(w, err.Error())
	}

	err = config.RunAnalysis()
	if err != nil {
		color.Red("Error: %v", err)
		health.Failure++
		fmt.Fprintf(w, err.Error())
	}

	if explain {
		err := config.GetAIResults(s.Output, anonymize)
		if err != nil {
			color.Red("Error: %v", err)
			health.Failure++
			fmt.Fprintf(w, err.Error())
		}
	}

	output, err := config.JsonOutput()
	if err != nil {
		color.Red("Error: %v", err)
		health.Failure++
		fmt.Fprintf(w, err.Error())
	}
	health.Success++
	fmt.Fprintf(w, string(output))
}

func (s *Config) Serve() error {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/analyze", s.analyzeHandler)
	http.HandleFunc("/healthz", s.healthzHandler)
	color.Green("Starting server on port %s", s.Port)
	err := http.ListenAndServe(":"+s.Port, nil)
	if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		return err
	}
	return nil
}

func (s *Config) healthzHandler(w http.ResponseWriter, r *http.Request) {
	js, err := json.MarshalIndent(health, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(js))
}

func getBoolParam(param string) bool {
	b, err := strconv.ParseBool(strings.ToLower(param))
	if err != nil {
		// Handle error if conversion fails
		return false
	}
	return b
}
