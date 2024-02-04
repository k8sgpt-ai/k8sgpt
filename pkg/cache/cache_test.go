package cache

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {

	cacheDir := "/tmp/cache"
	// Create cache directory if it doesn't exist.
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		err = os.Mkdir(cacheDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create cache directory: %v", err)
		}
	}

	// Write configuration to a file.
	configContent := []byte(`
        cache:
          file:
            directory: /tmp/cache
            disable: false
    `)
	err := ioutil.WriteFile(cacheDir+"/config.yaml", configContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(cacheDir)

	// Test the New function.
	cache := New("file")
	require.IsType(t, &FileBasedCache{}, cache)

	// Test the ParseCacheConfiguration function.
	cacheInfo, err := ParseCacheConfiguration()
	require.NoError(t, err)
	require.IsType(t, CacheProvider{}, cacheInfo)

	// Test the NewCacheProvider function.
	cacheProvider, err := NewCacheProvider("file", "", "", "", "", "")
	require.Error(t, err, "file is not a valid option")
	require.IsType(t, CacheProvider{}, cacheProvider)

	// Test the GetCacheConfiguration function.
	cache, err = GetCacheConfiguration()
	require.NoError(t, err)
	require.IsType(t, &FileBasedCache{}, cache)

	// Test the AddRemoteCache function.
	err = AddRemoteCache(CacheProvider{})
	require.NoError(t, err)

	// Test the RemoveRemoteCache function.
	err = RemoveRemoteCache()
	require.NoError(t, err)
}

