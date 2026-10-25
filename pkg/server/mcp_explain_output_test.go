package server

import (
	"os"
	"strings"
	"testing"
)

// TestMCPAnalyzeExplainPassesJSONOutput guards the fix for #1501:
// handleAnalyze must pass "json" (not an empty string) as the output
// format to GetAIResults, because GetAIResults only skips its
// progress bar when output == "json". The progress bar writes to
// stdout, which is the MCP stdio JSON-RPC channel — any progress
// output corrupts the protocol stream and the client drops the
// explained result, leaving "details" empty over MCP.
func TestMCPAnalyzeExplainPassesJSONOutput(t *testing.T) {
	src, err := os.ReadFile("mcp.go")
	if err != nil {
		t.Skipf("cannot read mcp.go from test environment: %v", err)
	}
	srcStr := string(src)
	want := `analysis.GetAIResults("json"`
	if !strings.Contains(srcStr, want) {
		t.Errorf("handleAnalyze must call GetAIResults with \"json\" (no `%s` in mcp.go); a non-json output makes GetAIResults start a progress bar on stdout, corrupting the MCP stdio JSON-RPC stream (issue #1501)", want)
	}
	if strings.Contains(srcStr, "var output string") {
		t.Error("handleAnalyze still declares `var output string` and passes it to GetAIResults; that empty output triggers the progress bar on stdout (issue #1501)")
	}
}
