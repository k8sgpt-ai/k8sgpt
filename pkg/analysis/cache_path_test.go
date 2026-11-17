package analysis

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/k8sgpt-ai/k8sgpt/pkg/ai"
	"github.com/k8sgpt-ai/k8sgpt/pkg/cache"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
	"github.com/stretchr/testify/require"
)

// cache.New falls back to FileBasedCache when the type is unrecognised, and
// that implementation writes real files under the user's XDG cache directory.
// Removing what a test put there keeps a CI runner from accumulating entries
// between runs, and stops one test's leftovers from being visible to another.
func cleanupKey(t *testing.T, c cache.ICache, key string) {
	t.Helper()
	t.Cleanup(func() { _ = c.Remove(key) })
}

func seedCache(t *testing.T, c cache.ICache, key, raw string) {
	t.Helper()
	require.NoError(t, c.Store(key, base64.StdEncoding.EncodeToString([]byte(raw))))
	require.True(t, c.Exists(key))
	cleanupKey(t, c, key)
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
	// the analysis regenerates and stores under newKey, so that entry needs
	// removing too, not just the one this test seeded
	cleanupKey(t, c, newKey)

	out, err := a.getAIResultForSanitizedFailures(texts, newTmpl)
	require.NoError(t, err)
	require.NotEqual(t, "STALE PRE-REWORD ANSWER", out)
	require.Equal(t, "I am a noop response to the prompt NEW reworded template in English: prompt-version-input", out)
}

// failingStoreCache is a file cache whose Store always fails.
type failingStoreCache struct{ cache.ICache }

func (failingStoreCache) Store(string, string) error { return errors.New("disk full") }

// The cache messages printed while `analyze --explain` fetches an AI result go
// to stderr, so stdout carries only the report.
func TestGetAIResult_CacheMessagesKeepStdoutClean(t *testing.T) {
	aiClient := &ai.NoOpAIClient{}
	texts := []string{"cache-message-input"}
	promptTmpl := "Respond in %s: %s"
	key := util.GetCacheKey(aiClient.GetName(), "English", promptTmpl+texts[0])

	tests := []struct {
		name  string
		cache func(t *testing.T) cache.ICache
		want  string
	}{
		{"empty entry", func(t *testing.T) cache.ICache {
			c := cache.New("message-empty-cache")
			seedCache(t, c, key, "   ")
			return c
		}, "cached data is empty; ignoring cache item"},
		{"undecodable entry", func(t *testing.T) cache.ICache {
			c := cache.New("message-decode-cache")
			require.NoError(t, c.Store(key, "not base64!"))
			cleanupKey(t, c, key)
			return c
		}, "error decoding cached data; ignoring cache item"},
		{"store failure", func(t *testing.T) cache.ICache {
			return failingStoreCache{cache.New("message-store-cache")}
		}, "error storing value to cache; value won't be cached: disk full"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Analysis{AIClient: aiClient, Cache: tt.cache(t), Language: "English"}
			var out string
			var err error
			stderr := captureStderr(t, func() {
				out, err = a.getAIResultForSanitizedFailures(texts, promptTmpl)
			})
			require.NoError(t, err)
			require.NotEmpty(t, out)
			require.Contains(t, stderr, tt.want)
		})
	}
}
