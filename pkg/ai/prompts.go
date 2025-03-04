package ai

const (
	default_prompt = `Simplify the following Kubernetes error message delimited by triple dashes written in --- %s --- language; --- %s ---.
	Provide the most possible solution in a step by step style in no more than 280 characters. Write the output in the following format:
	Error: {Explain error here}
	Solution: {Step by step solution here}
	`

	prom_conf_prompt = `Simplify the following Prometheus error message delimited by triple dashes written in --- %s --- language; --- %s ---.
	This error came when validating the Prometheus configuration file.
	Provide step by step instructions to fix, with suggestions, referencing Prometheus documentation if relevant.
	Write the output in the following format in no more than 300 characters:
	Error: {Explain error here}
	Solution: {Step by step solution here}
	`

	prom_relabel_prompt = `
	Return your prompt in this language: %s, beginning with
	The following is a list of the form:
	job_name:
	{Prometheus job_name}
	relabel_configs:
	{Prometheus relabel_configs}
	kubernetes_sd_configs:
	{Prometheus service discovery config}
	---
	%s
	---
	For each job_name, describe the Kubernetes service and pod labels,
	namespaces, ports, and containers they match.
	Return the message:
	Discovered and parsed Prometheus scrape configurations.
	For targets to be scraped by Prometheus, ensure they are running with
	at least one of the following label sets:
	Then for each job, write this format:
	- Job: {job_name}
	  - Service Labels:
	    - {list of service labels}
	  - Pod Labels:
	    - {list of pod labels}
	  - Namespaces:
	    - {list of namespaces}
	  - Ports:
	    - {list of ports}
	  - Containers:
	    - {list of container names}
	`

	kyverno_prompt = `Simplify the following Kyverno warnings message delimited by triple dashes written in --- %s --- language; --- %s ---.
	Provide the most probable solution as a kubectl command. 

	Write the output in the following format, for the solution, only show the kubectl command:
	
	Error: {Explain error here}

	Solution: {kubectl command}
	`
)

var PromptMap = map[string]string{
	"default":                       default_prompt,
	"PrometheusConfigValidate":      prom_conf_prompt,
	"PrometheusConfigRelabelReport": prom_relabel_prompt,
	"PolicyReport":                  kyverno_prompt,
	"ClusterPolicyReport":           kyverno_prompt,
}
