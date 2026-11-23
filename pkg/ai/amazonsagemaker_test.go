package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"                      //nolint:staticcheck // backend still uses aws-sdk-go v1
	"github.com/aws/aws-sdk-go/aws/credentials"          //nolint:staticcheck // backend still uses aws-sdk-go v1
	"github.com/aws/aws-sdk-go/aws/session"              //nolint:staticcheck // backend still uses aws-sdk-go v1
	"github.com/aws/aws-sdk-go/service/sagemakerruntime" //nolint:staticcheck // backend still uses aws-sdk-go v1
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The analysis layer already hands GetCompletion a complete instruction.
// The SageMaker body must carry that text as the user message and must not
// send the literal placeholder DEFAULT_PROMPT.
func TestSageMakerAIClient_GetCompletionSendsCallerPrompt(t *testing.T) {
	const prompt = "Respond ONLY in the --- English --- language. Simplify the following Kubernetes error."

	type captured struct {
		path string
		body []byte
	}
	got := make(chan captured, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		got <- captured{path: r.URL.Path, body: body}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"generation":{"role":"assistant","content":"Error: x\nSolution: y"}}]`))
	}))
	t.Cleanup(server.Close)

	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials("test", "test", ""),
		Endpoint:    aws.String(server.URL),
		Region:      aws.String("us-east-1"),
		DisableSSL:  aws.Bool(true),
	}))
	client := &SageMakerAIClient{
		client:      sagemakerruntime.New(sess),
		endpoint:    "k8sgpt-endpoint",
		temperature: 0.7,
		topP:        0.9,
		topK:        50,
		maxTokens:   256,
	}

	out, err := client.GetCompletion(context.Background(), prompt)
	require.NoError(t, err)
	assert.Equal(t, "Error: x\nSolution: y", out)

	req := <-got
	assert.Equal(t, "/endpoints/k8sgpt-endpoint/invocations", req.path)

	var body Request
	require.NoError(t, json.Unmarshal(req.body, &body))
	require.Len(t, body.Inputs, 1)
	require.Len(t, body.Inputs[0], 1)
	assert.Equal(t, "user", body.Inputs[0][0].Role)
	assert.Equal(t, prompt, body.Inputs[0][0].Content)
	for _, msg := range body.Inputs[0] {
		assert.NotEqual(t, "DEFAULT_PROMPT", msg.Content)
	}
	assert.Equal(t, 256, body.Parameters.MaxNewTokens)
	assert.InDelta(t, 0.9, body.Parameters.TopP, 0.0001)
	assert.InDelta(t, 50, body.Parameters.TopK, 0.0001)
	assert.InDelta(t, 0.7, body.Parameters.Temperature, 0.0001)
}
