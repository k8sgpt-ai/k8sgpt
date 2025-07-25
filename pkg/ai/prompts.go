package ai

const (
    // Prompt para erros genéricos do Kubernetes (ex: ImagePullBackOff)
    default_prompt = `Simplifique a seguinte mensagem de erro do Kubernetes delimitada por três traços escrita na linguagem --- %s ---; --- %s ---.
    Forneça a solução mais provável em um estilo passo a passo em não mais que 280 caracteres. Escreva a saída no seguinte formato:
    Error: {Explique o erro aqui}
    Solution: {Solução passo a passo aqui}
    `

	// Prompt para erros de configuração do Prometheus
    prom_conf_prompt = `Simplifique a seguinte mensagem de erro do Prometheus delimitada por três traços escrita na linguagem --- %s ---; --- %s ---.
    Este erro ocorreu ao validar o arquivo de configuração do Prometheus.
    Forneça instruções passo a passo para corrigir, com sugestões, referenciando a documentação do Prometheus se for relevante.
    Escreva a saída no seguinte formato, em não mais que 300 caracteres:
    Error: {Explique o erro aqui}
    Solution: {Solução passo a passo aqui}
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
    kyverno_prompt = `Simplifique a seguinte mensagem de aviso do Kyverno delimitada por três traços escrita na linguagem --- %s ---; --- %s ---.
    Forneça a solução mais provável como um comando kubectl.

    Escreva a saída no seguinte formato, para a solução, mostre apenas o comando kubectl:

    Error: {Explique o erro aqui}

    Solution: {comando kubectl}
    `
    // O prompt raw também precisa ser corrigido para corresponder à estrutura
    raw_promt = `{"language": "%s","message": "%s","prompt": "Simplifique a seguinte mensagem de erro do Kubernetes delimitada por três traços escrita na linguagem --- %s ---; --- %s ---. Forneça a solução mais provável em português."}`
)

var PromptMap = map[string]string{
	"raw":                           raw_promt,
	"default":                       default_prompt,
	"PrometheusConfigValidate":      prom_conf_prompt,
	"PrometheusConfigRelabelReport": prom_relabel_prompt,
	"PolicyReport":                  kyverno_prompt,
	"ClusterPolicyReport":           kyverno_prompt,
}
