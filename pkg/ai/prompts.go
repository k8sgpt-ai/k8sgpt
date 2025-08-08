package ai

const (
    // Prompt para erros genéricos do Kubernetes (ex: ImagePullBackOff)
    default_prompt = `Você é um assistente especializado em Kubernetes.
	Analise o seguinte erro e responda SOMENTE no formato JSON abaixo:

	{
	"error": "<breve descrição do erro>",
	"details": "<explicação detalhada sobre o que causou o problema>",
	"solution": "<passos claros para resolver>"
	}

	Erro: {{.error}}

	Idioma da resposta: português do Brasil.
	`

    // Prompt para erros de configuração do Prometheus
    prom_conf_prompt = `Explique o seguinte erro do Prometheus, delimitado por três traços, na linguagem --- %s ---: --- %s ---.
	Este erro ocorreu durante a validação do arquivo de configuração do Prometheus.
	Forneça instruções passo a passo para corrigir, em formato de lista numerada, com cada passo em uma nova linha, com sugestões, e referencie a documentação oficial do Prometheus se for relevante.
	Escreva a saída no seguinte formato, em não mais que 300 caracteres:
	Error: {Explique o erro aqui}
	Solution:
	{Liste os passos da solução aqui, formatados como uma lista numerada (ex: 1. Primeiro passo).}
`

    // Prompt para relatórios de relabeling do Prometheus
    prom_relabel_prompt = `
	Retorne o seu prompt neste idioma: %s, começando com
	A seguir está uma lista no formato:
	job_name:
	{nome_do_job_do_Prometheus}
	relabel_configs:
	{configurações de relabeling do Prometheus}
	kubernetes_sd_configs:
	{configuração de descoberta de serviço do Kubernetes}
	---
	%s
	---
	Para cada job_name, descreva os labels de serviço e pod do Kubernetes,
	namespaces, portas e contêineres que eles correspondem.
	Retorne a mensagem:
	Configurações de scrape do Prometheus descobertas e analisadas.
	Para que os alvos sejam scaneados pelo Prometheus, certifique-se de que estão
	sendo executados com pelo menos um dos seguintes conjuntos de labels:
	Em seguida, para cada job, escreva neste formato:
	- Job: {nome_do_job}
	- Labels de Serviço:
		- {lista de labels de serviço}
	- Labels de Pod:
		- {lista de labels de pod}
	- Namespaces:
		- {lista de namespaces}
	- Portas:
		- {lista de portas}
	- Contêineres:
		- {lista de nomes de contêineres}
	`

    // Prompt para avisos do Kyverno
    kyverno_prompt = `Explique o seguinte aviso do Kyverno, delimitado por três traços, na linguagem --- %s ---: --- %s ---.
	Forneça a solução mais provável como um comando kubectl. Escreva a saída no seguinte formato:
	Error: {Explique o erro aqui}
	Solution:
	{comando kubectl}
	`

    // Prompt raw para backend google/gemini, já com JSON puro e idioma fixo em português do Brasil
    raw_promt = `{"language": "%s","message": "%s","prompt": "Analise o seguinte erro do Kubernetes, delimitado por três traços, na linguagem --- %s ---: --- %s ---. Responda somente no formato JSON sem blocos de código, no idioma português do Brasil: {\"error\":\"<breve descrição>\",\"details\":\"<explicação detalhada>\",\"solution\":\"<passos claros para resolver>\"}"}`
)

var PromptMap = map[string]string{
    "raw":                           raw_promt,
    "default":                       default_prompt,
    "PrometheusConfigValidate":      prom_conf_prompt,
    "PrometheusConfigRelabelReport": prom_relabel_prompt,
    "PolicyReport":                  kyverno_prompt,
    "ClusterPolicyReport":           kyverno_prompt,
}
