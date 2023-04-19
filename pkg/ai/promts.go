package ai

const (
	default_prompt = "Simplify the following Kubernetes error message and provide a solution in %s: %s"
	prompt_a       = "Read the following input %s and provide possible scenarios for remediation in %s"
	prompt_b       = "Considering the following input from the Kubernetes resource %s and the error message %s, provide possible scenarios for remediation in %s"
	prompt_c       = "Reading the following %s error message and it's accompanying log message %s, how would you simplify this message?"
)
