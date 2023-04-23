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
	Port           string
	Backend        string
	Key            string
	Token          string
	Output         string
	maxConcurrency int
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

	var err error
	s.maxConcurrency, err = strconv.Atoi(r.URL.Query().Get("maxConcurrency"))
	if err != nil {
		s.maxConcurrency = 10
	}
	s.Output = r.URL.Query().Get("output")

	if s.Output == "" {
		s.Output = "json"
	}

	config, err := analysis.NewAnalysis(s.Backend, language, []string{}, namespace, nocache, explain, s.maxConcurrency)
	if err != nil {
		health.Failure++
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	analysisErrors := config.RunAnalysis()
	if analysisErrors != nil {
		var errorMessage string
		for _, err := range analysisErrors {
			errorMessage += err.Error() + "\n"
		}
		http.Error(w, errorMessage, http.StatusInternalServerError)
		health.Failure++
	}

	if explain {
		err := config.GetAIResults(s.Output, anonymize)
		if err != nil {
			health.Failure++
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	out, err := config.PrintOutput(s.Output)
	if err != nil {
		health.Failure++
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	health.Success++
	fmt.Fprintf(w, string(out))
}

func (s *Config) Serve() error {
	handler := loggingMiddleware(http.DefaultServeMux)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/analyze", s.analyzeHandler)
	http.HandleFunc("/healthz", s.healthzHandler)
	color.Green("Starting server on port %s", s.Port)
	err := http.ListenAndServe(":"+s.Port, handler)
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
