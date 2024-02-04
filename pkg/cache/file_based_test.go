package cache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileBasedCache(t *testing.T) {
	// Create a new FileBasedCache.
	f := &FileBasedCache{}

	// Test the Configure method.
	err := f.Configure(CacheProvider{})
	require.NoError(t, err)

	// Test the IsCacheDisabled method.
	require.False(t, f.IsCacheDisabled())

	// Test the Store method.
	err = f.Store("test-key", "test-data")
	require.NoError(t, err)

	// Test the Load method.
	data, err := f.Load("test-key")
	require.NoError(t, err)
	require.Equal(t, "test-data", data)

	// Test the Exists method.
	exists := f.Exists("test-key")
	require.True(t, exists)

	// Test the not exists case.
	exists = f.Exists("test-key-not-exists")
	require.False(t, exists)

	// Test the List method.
	list, err := f.List()
	require.NoError(t, err)
	require.Greater(t, len(list), 0)

	// Test the Remove method.
	err = f.Remove("test-key")
	require.NoError(t, err)

	// Test the Exists method again.
	exists = f.Exists("test-key")
	require.False(t, exists)

	// Test the GetName method.
	name := f.GetName()
	require.Equal(t, "file", name)

	// Test the DisableCache method.
	f.DisableCache()
	require.True(t, f.IsCacheDisabled())
}
