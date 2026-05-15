# Supported AI Providers and Models in K8sGPT

K8sGPT supports a variety of AI/LLM providers (backends). Some providers have a fixed set of supported models, while others allow you to specify any model supported by the provider.

---

## Providers and Supported Models

### OpenAI
- **Model:** User-configurable (any model supported by OpenAI, e.g., `gpt-3.5-turbo`, `gpt-4`, etc.)

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

### Amazon Bedrock Converse
- **Model:** User-configurable (any model supported by [Amazon Bedrock Converse](https://docs.aws.amazon.com/bedrock/latest/userguide/models-api-compatibility.html))

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

### Amazon Bedrock Mantle (OpenAI-compatible)
- **Auth:** Reads the API key from the `AWS_BEARER_TOKEN_BEDROCK` environment variable
- **Region:** Required — the endpoint is auto-constructed as `https://bedrock-mantle.{region}.api.aws/v1`
- **API:** Uses the OpenAI-compatible Chat Completions API (`/v1/chat/completions`)
- **Supported Models:**

  | Provider | Models |
  |---|---|
  | Anthropic | Claude Mythos Preview, Claude Opus 4.7 |
  | DeepSeek | DeepSeek V3.2, DeepSeek-V3.1 |
  | Google | Gemma 3 4B IT, Gemma 3 12B IT, Gemma 3 27B PT |
  | MiniMax | MiniMax M2, M2.1, M2.5 |
  | Mistral AI | Devstral 2 123B, Magistral Small 2509, Ministral 14B 3.0, Ministral 3 8B, Ministral 3B, Mistral Large 3, Voxtral Mini 3B, Voxtral Small 24B |
  | Moonshot AI | Kimi K2 Thinking, Kimi K2.5 |
  | NVIDIA | Nemotron Nano 9B v2, Nemotron Nano 12B v2 VL, Nemotron Nano 3 30B, Nemotron 3 Super 120B |
  | OpenAI | gpt-oss-120b, gpt-oss-20b, GPT OSS Safeguard 120B, GPT OSS Safeguard 20B |
  | Qwen | Qwen3 235B, Qwen3 32B, Qwen3 Coder 480B, Qwen3 Coder Next, Qwen3 Next 80B, Qwen3 VL 235B, Qwen3-Coder-30B |
  | Writer | Palmyra Vision 7B |
  | Z.AI | GLM 4.7, GLM 4.7 Flash, GLM 5 |

  > **Note:** For the latest model availability, see [Endpoint availability](https://docs.aws.amazon.com/bedrock/latest/userguide/models-endpoint-availability.html).

---

For more details on configuring each provider and model, refer to the official K8sGPT documentation and the provider's own documentation. 
