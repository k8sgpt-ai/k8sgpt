package analysis

import (
	"encoding/base64"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/cache"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	"github.com/stretchr/testify/require"
)

func seedCache(t *testing.T, c cache.ICache, key, raw string) {
	t.Helper()
	require.NoError(t, c.Store(key, base64.StdEncoding.EncodeToString([]byte(raw))))
	require.True(t, c.Exists(key))
}

// Cache-hit empty guard: a cached-but-empty entry must be treated as a miss
// and regenerated, not returned as a false-success empty result.
func TestGetAIResult_EmptyCacheHitFallsThrough(t *testing.T) {
	aiClient := &ai.NoOpAIClient{}
	c := cache.New("empty-guard-cache")
	a := Analysis{AIClient: aiClient, Cache: c, Language: "English"}

	texts := []string{"empty-guard-input"}
	promptTmpl := "Respond in %s: %s"
	inputKey := texts[0]
	key := util.GetCacheKey(aiClient.GetName(), a.Language, promptTmpl+inputKey)

	seedCache(t, c, key, "   ")

	out, err := a.getAIResultForSanitizedFailures(texts, promptTmpl)
	require.NoError(t, err)
	require.NotEmpty(t, out)
	require.Equal(t, "I am a noop response to the prompt Respond in English: empty-guard-input", out)
}

// Non-empty cache hit is still served directly (guard does not break the happy path).
func TestGetAIResult_NonEmptyCacheHitReturned(t *testing.T) {
	aiClient := &ai.NoOpAIClient{}
	c := cache.New("nonempty-hit-cache")
	a := Analysis{AIClient: aiClient, Cache: c, Language: "English"}

	texts := []string{"nonempty-hit-input"}
	promptTmpl := "Respond in %s: %s"
	inputKey := texts[0]
	key := util.GetCacheKey(aiClient.GetName(), a.Language, promptTmpl+inputKey)

	seedCache(t, c, key, "a real cached answer")

	out, err := a.getAIResultForSanitizedFailures(texts, promptTmpl)
	require.NoError(t, err)
	require.Equal(t, "a real cached answer", out)
}

// Prompt-versioned key: an entry cached under an OLD prompt template must not be
// served when the template wording changes; it regenerates instead of serving stale.
func TestGetAIResult_PromptVersionedKeyInvalidatesStale(t *testing.T) {
	aiClient := &ai.NoOpAIClient{}
	c := cache.New("prompt-version-cache")
	a := Analysis{AIClient: aiClient, Cache: c, Language: "English"}

	texts := []string{"prompt-version-input"}
	inputKey := texts[0]

	oldTmpl := "OLD template in %s: %s"
	newTmpl := "NEW reworded template in %s: %s"

	staleKey := util.GetCacheKey(aiClient.GetName(), a.Language, oldTmpl+inputKey)
	seedCache(t, c, staleKey, "STALE PRE-REWORD ANSWER")

	newKey := util.GetCacheKey(aiClient.GetName(), a.Language, newTmpl+inputKey)
	require.NotEqual(t, staleKey, newKey)

	out, err := a.getAIResultForSanitizedFailures(texts, newTmpl)
	require.NoError(t, err)
	require.NotEqual(t, "STALE PRE-REWORD ANSWER", out)
	require.Equal(t, "I am a noop response to the prompt NEW reworded template in English: prompt-version-input", out)
}
