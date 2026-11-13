package ai

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// A hand-written provider omits maxtokens, so GetMaxTokens() is 0. Bedrock
// rejects a request whose max_tokens, maxTokenCount or max_new_tokens is 0.
// The per-model catalog already carries a positive limit; that value has to
// go out unless the user set one.
func TestAmazonBedrockUnsetMaxTokensKeepsCatalogDefault(t *testing.T) {
	cases := []struct {
		name       string
		model      string
		configured int
		reply      string
		path       []string
	}{
		{
			name:  "messages max_tokens",
			model: "anthropic.claude-sonnet-4-20250514-v1:0",
			reply: `{"content":[{"type":"text","text":"ok"}]}`,
			path:  []string{"max_tokens"},
		},
		{
			name:  "titan maxTokenCount",
			model: "amazon.titan-text-express-v1",
			reply: `{"results":[{"outputText":"ok"}]}`,
			path:  []string{"textGenerationConfig", "maxTokenCount"},
		},
		{
			name:  "nova max_new_tokens",
			model: "amazon.nova-pro-v1:0",
			reply: `{"output":{"message":{"role":"assistant","content":[{"text":"ok"}]}}}`,
			path:  []string{"inferenceConfig", "max_new_tokens"},
		},
		{
			name:       "configured max tokens overrides the catalog",
			model:      "anthropic.claude-sonnet-4-20250514-v1:0",
			configured: 512,
			reply:      `{"content":[{"type":"text","text":"ok"}]}`,
			path:       []string{"max_tokens"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			catalog := bedrockCatalogMaxTokens(t, tc.model)
			want := catalog
			if tc.configured > 0 {
				want = tc.configured
			}

			mockRuntime := new(MockBedrockRuntimeClient)
			var body []byte
			mockRuntime.On("InvokeModel", mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					in := args.Get(1).(*bedrockruntime.InvokeModelInput)
					body = append([]byte(nil), in.Body...)
				}).
				Return(&bedrockruntime.InvokeModelOutput{Body: []byte(tc.reply)}, nil)

			client := NewAmazonBedRockClient(nil)
			client.client = mockRuntime
			client.mgmtClient = new(MockBedrockClient)

			err := client.Configure(&AIProvider{
				Model:          tc.model,
				ProviderRegion: "us-east-1",
				MaxTokens:      tc.configured,
			})
			assert.NoError(t, err)

			out, err := client.GetCompletion(context.Background(), "explain this")
			assert.NoError(t, err)
			assert.Equal(t, "ok", out)
			assert.Equal(t, want, bedrockRequestInt(t, body, tc.path...))
			mockRuntime.AssertExpectations(t)
		})
	}
}

func bedrockCatalogMaxTokens(t *testing.T, model string) int {
	t.Helper()
	for _, m := range defaultModels {
		if m.Name == model {
			assert.Greater(t, m.Config.MaxTokens, 0)
			return m.Config.MaxTokens
		}
	}
	t.Fatalf("model %s is not in the bedrock catalog", model)
	return 0
}

func bedrockRequestInt(t *testing.T, body []byte, path ...string) int {
	t.Helper()
	var cur interface{}
	if !assert.NoError(t, json.Unmarshal(body, &cur)) {
		return -1
	}
	for _, key := range path {
		obj, ok := cur.(map[string]interface{})
		if !assert.Truef(t, ok, "expected object before %s in %s", key, body) {
			return -1
		}
		cur, ok = obj[key]
		if !assert.Truef(t, ok, "missing %s in %s", key, body) {
			return -1
		}
	}
	n, ok := cur.(float64)
	if !assert.Truef(t, ok, "expected number at %v, got %T (%v)", path, cur, cur) {
		return -1
	}
	return int(n)
}
