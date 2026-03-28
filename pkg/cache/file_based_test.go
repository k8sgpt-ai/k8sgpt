package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/require"
)

// withTempCacheHome sets XDG_CACHE_HOME to a temp dir for test isolation.
func withTempCacheHome(t *testing.T) func() {
	t.Helper()
	tmp, err := os.MkdirTemp("", "k8sgpt-cache-test-*")
	require.NoError(t, err)
	old := os.Getenv("XDG_CACHE_HOME")
	require.NoError(t, os.Setenv("XDG_CACHE_HOME", tmp))
	return func() {
		_ = os.Setenv("XDG_CACHE_HOME", old)
		_ = os.RemoveAll(tmp)
	}
}

func TestFileBasedCache_BasicOps(t *testing.T) {
	cleanup := withTempCacheHome(t)
	defer cleanup()

	c := &FileBasedCache{}
	// Configure should be a no-op
	require.NoError(t, c.Configure(CacheProvider{}))
	require.Equal(t, "file", c.GetName())
	require.False(t, c.IsCacheDisabled())
	c.DisableCache()
	require.True(t, c.IsCacheDisabled())

	key := "testkey"
	data := "hello"

	// Store
	require.NoError(t, c.Store(key, data))

	// Exists
	require.True(t, c.Exists(key))

	// Load
	got, err := c.Load(key)
	require.NoError(t, err)
	require.Equal(t, data, got)

	// List should include our key file
	items, err := c.List()
	require.NoError(t, err)
	// ensure at least one item and that one matches our key
	found := false
	for _, it := range items {
		if it.Name == key {
			found = true
			break
		}
	}
	require.True(t, found)

	// Remove
	require.NoError(t, c.Remove(key))
	require.False(t, c.Exists(key))
}

func TestFileBasedCache_PathShape(t *testing.T) {
	cleanup := withTempCacheHome(t)
	defer cleanup()
	// Verify xdg.CacheFile path shape (directory and filename)
	p, err := xdg.CacheFile(filepath.Join("k8sgpt", "abc"))
	require.NoError(t, err)
	require.Equal(t, "abc", filepath.Base(p))
	require.Contains(t, p, "k8sgpt")
}
