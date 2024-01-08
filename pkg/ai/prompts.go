package ai

import "text/template"

const (
	trivy_vuln_prompt = "Explain the following trivy scan result and the detail risk or root cause of the CVE ID, then provide a solution. Response in %s: %s"
	trivy_conf_prompt = "Explain the following trivy scan result and the detail risk or root cause of the security check, then provide a solution."
)

var PromptMap = map[string]*template.Template{
	"default": template.Must(template.New("default").Parse(
		`Simplify the following Kubernetes error message delimited by triple dashes written in --- {{ .Failure.Text }} --- language; --- {{ .Language }} ---.

{{ .Failure.AdditionalContextText }}
Provide the most possible solution in a step by step style in no more than 280 characters. Write the output in the following format:
Error: {Explain error here}
Solution: {Step by step solution here}`,
	)),

	// TODO(bwplotka): Fix
	//"VulnerabilityReport": trivy_vuln_prompt, // for Trivy integration, the key should match `Result.Kind` in pkg/common/types.go
	//"ConfigAuditReport":   trivy_conf_prompt,
}
