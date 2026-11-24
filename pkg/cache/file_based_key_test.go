package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/require"
)

// isolateXDGCache points XDG_CACHE_HOME at a fresh directory inside root and
// reloads xdg, which reads the variable only at init or on Reload.
func isolateXDGCache(t *testing.T, root string) {
	t.Helper()
	t.Setenv("XDG_CACHE_HOME", filepath.Join(root, "cache"))
	xdg.Reload()
	t.Cleanup(xdg.Reload)
}

func TestFileBasedCache_RejectsKeysOutsideCacheDir(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "parent directory", key: ".."},
		{name: "parent traversal", key: "../outside"},
		{name: "deep traversal", key: "../../outside"},
		{name: "traversal after a subdirectory", key: "sub/../../outside"},
		{name: "absolute path", key: "/tmp/outside"},
		{name: "empty key", key: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			isolateXDGCache(t, root)

			// A file just outside the cache directory that an escaping key
			// could otherwise read or delete.
			outside := filepath.Join(root, "outside")
			require.NoError(t, os.WriteFile(outside, []byte("keep"), 0600))

			c := &FileBasedCache{}

			require.Error(t, c.Store(tt.key, "data"))
			_, err := c.Load(tt.key)
			require.Error(t, err)
			require.False(t, c.Exists(tt.key))
			require.Error(t, c.Remove(tt.key))

			got, err := os.ReadFile(outside)
			require.NoError(t, err)
			require.Equal(t, "keep", string(got))

			// The check runs before xdg.CacheFile, so no directories are
			// created for a rejected key.
			_, err = os.Stat(filepath.Join(root, "cache"))
			require.True(t, os.IsNotExist(err), "cache directory should not be created for key %q", tt.key)
		})
	}
}

func TestFileBasedCache_AcceptsLocalKeys(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "hash key", key: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"},
		{name: "plain name", key: "testkey"},
		{name: "dots inside the name", key: "a..b"},
		{name: "subdirectory", key: "sub/key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			isolateXDGCache(t, root)

			c := &FileBasedCache{}

			require.NoError(t, c.Store(tt.key, "data"))
			require.True(t, c.Exists(tt.key))
			got, err := c.Load(tt.key)
			require.NoError(t, err)
			require.Equal(t, "data", got)

			_, err = os.Stat(filepath.Join(root, "cache", "k8sgpt", tt.key))
			require.NoError(t, err)

			require.NoError(t, c.Remove(tt.key))
			require.False(t, c.Exists(tt.key))
		})
	}
}
