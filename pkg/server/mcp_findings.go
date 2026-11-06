/*
Copyright 2026 The K8sGPT Authors.
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
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/k8sgpt-ai/k8sgpt/pkg/analysis"
	"github.com/k8sgpt-ai/k8sgpt/pkg/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/spf13/viper"
)

// FindingsSchemaVersion is the version of the FindingsReport document. Bump it
// on any breaking change so MCP clients can detect the shape they receive.
const FindingsSchemaVersion = "k8sgpt.ai/findings/v1alpha1"

const (
	findingsResourceURI         = "k8sgpt://findings"
	findingsResourceTemplateURI = "k8sgpt://findings/{namespace}"
)

// FindingsReport is the structured, AI-free result of running the analyzers.
// It is meant for programs (agents, MCP clients), not people: every finding is
// addressable by kind, namespace and name, and carries only the analyzer's
// own failure messages.
type FindingsReport struct {
	SchemaVersion string    `json:"schemaVersion"`
	GeneratedAt   time.Time `json:"generatedAt"`
	Namespace     string    `json:"namespace,omitempty"`
	Filters       []string  `json:"filters,omitempty"`
	Problems      int       `json:"problems"`
	Findings      []Finding `json:"findings"`
	Errors        []string  `json:"errors,omitempty"`
}

// Finding is one analyzer result.
type Finding struct {
	Kind         string   `json:"kind"`
	Namespace    string   `json:"namespace,omitempty"`
	Name         string   `json:"name"`
	ParentObject string   `json:"parentObject,omitempty"`
	Errors       []string `json:"errors"`
}

// findingsRequest holds the arguments of the get-findings tool.
type findingsRequest struct {
	Namespace     string   `json:"namespace,omitempty"`
	Filters       []string `json:"filters,omitempty"`
	LabelSelector string   `json:"labelSelector,omitempty"`
}

// runFindingsAnalysis runs the analyzers without an AI backend. It is a
// variable so tests can replace it.
var runFindingsAnalysis = func(namespace string, filters []string, labelSelector string) ([]common.Result, []string, error) {
	a, err := analysis.NewAnalysis(
		"",            // backend: unused without explain
		"english",     // language: unused without explain
		filters,       // filters
		namespace,     // namespace
		labelSelector, // labelSelector
		true,          // noCache: there is nothing to cache without explain
		false,         // explain: findings never call an AI backend
		validateMaxConcurrency(0),
		false, // withDoc
		false, // interactiveMode
		nil,   // httpHeaders
		false, // withStats
	)
	if err != nil {
		return nil, nil, err
	}
	defer a.Close()
	a.RunAnalysis()
	return a.Results, a.Errors, nil
}

// buildFindingsReport converts analyzer results into a FindingsReport with a
// stable order, so identical cluster state gives identical documents.
func buildFindingsReport(namespace string, filters []string, results []common.Result, errs []string, now time.Time) FindingsReport {
	findings := make([]Finding, 0, len(results))
	for _, r := range results {
		f := Finding{Kind: r.Kind, Name: r.Name, ParentObject: r.ParentObject, Errors: make([]string, 0, len(r.Error))}
		if ns, name, ok := strings.Cut(r.Name, "/"); ok {
			f.Namespace, f.Name = ns, name
		}
		for _, failure := range r.Error {
			f.Errors = append(f.Errors, failure.Text)
		}
		findings = append(findings, f)
	}
	sort.Slice(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Namespace != b.Namespace {
			return a.Namespace < b.Namespace
		}
		return a.Name < b.Name
	})
	return FindingsReport{
		SchemaVersion: FindingsSchemaVersion,
		GeneratedAt:   now.UTC(),
		Namespace:     namespace,
		Filters:       filters,
		Problems:      len(findings),
		Findings:      findings,
		Errors:        errs,
	}
}

func (s *K8sGptMCPServer) findingsJSON(namespace string, filters []string, labelSelector string) (string, error) {
	if len(filters) == 0 {
		filters = viper.GetStringSlice("active_filters")
	}
	results, errs, err := runFindingsAnalysis(namespace, filters, labelSelector)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(buildFindingsReport(namespace, filters, results, errs, time.Now()))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// registerFindings registers the get-findings tool and the findings
// resources.
func (s *K8sGptMCPServer) registerFindings() {
	findingsTool := mcp.NewTool("get-findings",
		mcp.WithDescription("Run the K8sGPT analyzers and return their findings as structured JSON ("+FindingsSchemaVersion+"). Never calls an AI backend; use the analyze tool with explain for explanations."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("namespace",
			mcp.Description("Namespace to analyze; empty analyzes all namespaces"),
		),
		mcp.WithArray("filters",
			mcp.Description("Analyzers to run (e.g. ['Pod', 'Deployment']); empty uses the active filters"),
			mcp.WithStringItems(),
		),
		mcp.WithString("labelSelector",
			mcp.Description("Label selector to narrow down the analyzed resources (e.g. app=nginx)"),
		),
	)
	s.server.AddTool(findingsTool, s.handleGetFindings)

	s.server.AddResource(mcp.NewResource(findingsResourceURI, "findings",
		mcp.WithResourceDescription("Current K8sGPT findings for the whole cluster, using the active filters"),
		mcp.WithMIMEType("application/json"),
	), s.readFindingsResource)

	s.server.AddResourceTemplate(mcp.NewResourceTemplate(findingsResourceTemplateURI, "namespace-findings",
		mcp.WithTemplateDescription("Current K8sGPT findings for one namespace, using the active filters"),
		mcp.WithTemplateMIMEType("application/json"),
	), s.readFindingsResource)
}

func (s *K8sGptMCPServer) handleGetFindings(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var req findingsRequest
	if err := request.BindArguments(&req); err != nil {
		return mcp.NewToolResultErrorf("Failed to parse request arguments: %v", err), nil
	}
	out, err := s.findingsJSON(req.Namespace, req.Filters, req.LabelSelector)
	if err != nil {
		return mcp.NewToolResultErrorf("Failed to run analysis: %v", err), nil
	}
	return mcp.NewToolResultText(out), nil
}

// readFindingsResource serves both k8sgpt://findings and
// k8sgpt://findings/{namespace}.
func (s *K8sGptMCPServer) readFindingsResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	uri := request.Params.URI
	namespace := ""
	if rest, ok := strings.CutPrefix(uri, findingsResourceURI+"/"); ok {
		namespace = rest
		if namespace == "" || strings.Contains(namespace, "/") {
			return nil, fmt.Errorf("invalid findings resource URI %q", uri)
		}
	} else if uri != findingsResourceURI {
		return nil, fmt.Errorf("unknown findings resource URI %q", uri)
	}
	out, err := s.findingsJSON(namespace, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to run analysis: %w", err)
	}
	return []mcp.ResourceContents{
		&mcp.TextResourceContents{URI: uri, MIMEType: "application/json", Text: out},
	}, nil
}
