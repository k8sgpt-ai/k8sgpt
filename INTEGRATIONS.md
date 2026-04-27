# Integrations

k8sgpt integrates with a variety of cloud native tools, platforms, and services.

## CNCF Project Integrations

| Project | Integration Type | Description |
|---------|-----------------|-------------|
| [Prometheus](https://prometheus.io/) | Exporter / Metrics | k8sgpt operator can export analysis results to Prometheus for monitoring |
| [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator) | Operator Integration | Integration with Prometheus Operator for service discovery and alerting |
| [Alertmanager](https://prometheus.io/docs/alerting/alertmanager/) | Alert Integration | Send k8sgpt analysis alerts to Alertmanager |
| [OpenTelemetry](https://opentelemetry.io/) | Observability | Export analysis metrics and traces via OpenTelemetry |
| [Grafana](https://grafana.com/) | Dashboard | Visualize k8sgpt analysis results in Grafana dashboards |
| [Kubernetes](https://kubernetes.io/) | Core Platform | Native Kubernetes resource analysis and diagnostics |
| [Helm](https://helm.sh/) | Packaging | k8sgpt available as a Helm chart in the [charts](https://github.com/k8sgpt-ai/charts) repository |
| [Krew](https://krew.dev/) | Plugin Distribution | k8sgpt distributed as a Krew kubectl plugin via [.krew.yaml](https://github.com/k8sgpt-ai/k8sgpt/blob/main/.krew.yaml) |

## AI/LLM Provider Integrations

| Provider | Backend Name | Description |
|----------|-------------|-------------|
| [OpenAI](https://openai.com/) | `openai` | Default provider - supports GPT-3.5, GPT-4, and other OpenAI models |
| [Azure OpenAI](https://azure.microsoft.com/services/openai/) | `azureopenai` | Azure-hosted OpenAI models |
| [Cohere](https://cohere.com/) | `cohere` | Cohere's command models |
| [Amazon Bedrock](https://aws.amazon.com/bedrock/) | `amazonbedrock` | AWS Bedrock - supports Claude, Llama, Titan, and more |
| [Amazon SageMaker](https://aws.amazon.com/sagemaker/) | `amazonsagemaker` | AWS SageMaker JumpStart models |
| [Google Gemini](https://ai.google/gemini) | `google` | Google's Gemini models |
| [Google Vertex AI](https://cloud.google.com/vertex-ai) | `googlevertexai` | Google Cloud Vertex AI models |
| [Ollama](https://ollama.com/) | `ollama` | Local LLM inference with Ollama |
| [LocalAI](https://localai.io/) | `localai` | Self-hosted OpenAI-compatible API |
| [Hugging Face](https://huggingface.co/) | `huggingface` | Hugging Face Inference API |
| [IBM WatsonX](https://www.ibm.com/watsonx) | `watsonxai` | IBM WatsonX AI models |
| [IBM WatsonxAI](https://www.ibm.com/watsonx) | `ibmwatsonxai` | IBM WatsonxAI specific integration |
| [Custom REST](https://) | `customrest` | Any REST API that follows the OpenAI chat completion format |

## Other Tool Integrations

| Tool | Integration Type | Description |
|------|-----------------|-------------|
| [Claude Desktop](https://claude.ai/) | MCP Server | k8sgpt MCP server integrates with Claude Desktop for AI-assisted cluster analysis |
| [Docker](https://www.docker.com/) | Container | Container image available on GitHub Container Registry |
| [Minikube](https://minikube.sigs.k8s.io/) | Development | Works with Minikube clusters for development and testing |
| [Kubeblocks](https://kubeblocks.io/) | Database Analysis | Analyzer support for KubeBlocks-managed databases |
| [OpenShift](https://www.openshift.com/) | Platform | Analysis support for OpenShift-specific resources (CatalogSource, ClusterCatalog, etc.) |
| [FluxCD](https://fluxcd.io/) | GitOps | Compatible with GitOps workflows using FluxCD |
| [ArgoCD](https://argoproj.github.io/cd/) | GitOps | Compatible with GitOps workflows using ArgoCD |

## Remote Caching

k8sgpt supports remote caching of analysis results:

| Provider | Type | Description |
|----------|------|-------------|
| AWS S3 | Object Storage | Store analysis cache in AWS S3 buckets |
| Azure Blob | Object Storage | Store analysis cache in Azure Blob Storage |
| Google Cloud Storage | Object Storage | Store analysis cache in GCS buckets |

## How to Integrate

For custom analyzer integrations, see the [Custom Analyzers documentation](https://docs.k8sgpt.ai/tutorials/custom-analyzers/) and the [custom analyzer schema](https://github.com/k8sgpt-ai/schemas/blob/main/protobuf/schema/v1/custom_analyzer.proto).

For MCP server integration, see [MCP.md](MCP.md).
