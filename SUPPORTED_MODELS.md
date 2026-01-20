# Supported AI Providers and Models in K8sGPT

K8sGPT supports a variety of AI/LLM providers (backends). Some providers have a fixed set of supported models, while others allow you to specify any model supported by the provider.

---

## Providers and Supported Models

### OpenAI
- **Model:** User-configurable (any model supported by OpenAI, e.g., `gpt-3.5-turbo`, `gpt-4`, etc.)

### OpenAI Responses
- **Model:** User-configurable (any model supported by the OpenAI Responses API, e.g., `gpt-4.1`, `o1`, `gpt-5`)

### Azure OpenAI
- **Model:** User-configurable (any model deployed in your Azure OpenAI resource)

### LocalAI
- **Model:** User-configurable (default: `llama3`)

### Ollama
- **Model:** User-configurable (default: `llama3`, others can be specified)

### NoOpAI
- **Model:** N/A (no real model, used for testing)

### Cohere
- **Model:** User-configurable (any model supported by Cohere)

### Amazon Bedrock
- **Supported Models:**
  - anthropic.claude-sonnet-4-20250514-v1:0
  - us.anthropic.claude-sonnet-4-20250514-v1:0
  - eu.anthropic.claude-sonnet-4-20250514-v1:0
  - apac.anthropic.claude-sonnet-4-20250514-v1:0
  - us.anthropic.claude-3-7-sonnet-20250219-v1:0
  - eu.anthropic.claude-3-7-sonnet-20250219-v1:0
  - apac.anthropic.claude-3-7-sonnet-20250219-v1:0
  - anthropic.claude-3-5-sonnet-20240620-v1:0
  - us.anthropic.claude-3-5-sonnet-20241022-v2:0
  - anthropic.claude-v2
  - anthropic.claude-v1
  - anthropic.claude-instant-v1
  - ai21.j2-ultra-v1
  - ai21.j2-jumbo-instruct
  - amazon.titan-text-express-v1
  - amazon.nova-pro-v1:0
  - eu.amazon.nova-pro-v1:0
  - us.amazon.nova-pro-v1:0
  - amazon.nova-lite-v1:0
  - eu.amazon.nova-lite-v1:0
  - us.amazon.nova-lite-v1:0
  - anthropic.claude-3-haiku-20240307-v1:0

> **Note:**
> If you use an AWS Bedrock inference profile ARN (e.g., `arn:aws:bedrock:us-east-1:<account>:application-inference-profile/<id>`) as the model, you must still provide a valid modelId (e.g., `anthropic.claude-3-sonnet-20240229-v1:0`). K8sGPT will automatically set the required `X-Amzn-Bedrock-Inference-Profile-ARN` header for you when making requests to Bedrock.

### Amazon SageMaker
- **Model:** User-configurable (any model deployed in your SageMaker endpoint)

### Google GenAI
- **Model:** User-configurable (any model supported by Google GenAI, e.g., `gemini-pro`)

### Huggingface
- **Model:** User-configurable (any model supported by Huggingface Inference API)

### Google VertexAI
- **Supported Models:**
  - gemini-1.0-pro-001

### OCI GenAI
- **Model:** User-configurable (any model supported by OCI GenAI)

### Custom REST
- **Model:** User-configurable (any model your custom REST endpoint supports)

### IBM Watsonx
- **Supported Models:**
  - ibm/granite-13b-chat-v2

### Groq
- **Model:** User-configurable (any model supported by Groq, e.g., `llama-3.3-70b-versatile`, `mixtral-8x7b-32768`)

---

For more details on configuring each provider and model, refer to the official K8sGPT documentation and the provider's own documentation. 
