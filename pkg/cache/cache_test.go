package cache

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestNewReturnsExpectedCache(t *testing.T) {
	require.IsType(t, &FileBasedCache{}, New("file"))
	require.IsType(t, &AzureCache{}, New("azure"))
	require.IsType(t, &GCSCache{}, New("gcs"))
	require.IsType(t, &S3Cache{}, New("s3"))
	require.IsType(t, &InterplexCache{}, New("interplex"))
	// default fallback
	require.IsType(t, &FileBasedCache{}, New("unknown"))
}

func TestNewCacheProvider_InterplexAndInvalid(t *testing.T) {
	// valid: interplex
	cp, err := NewCacheProvider("interplex", "", "", "localhost:1", "", "", "", false)
	require.NoError(t, err)
	require.Equal(t, "interplex", cp.CurrentCacheType)
	require.Equal(t, "localhost:1", cp.Interplex.ConnectionString)

	// invalid type
	_, err = NewCacheProvider("not-a-type", "", "", "", "", "", "", false)
	require.Error(t, err)
}

func TestAddRemoveRemoteCacheAndGet(t *testing.T) {
	// isolate viper with temp config file
	tmpFile, err := os.CreateTemp("", "k8sgpt-cache-config-*.yaml")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()
	viper.Reset()
	viper.SetConfigFile(tmpFile.Name())

	// add interplex remote cache
	cp := CacheProvider{}
	cp.CurrentCacheType = "interplex"
	cp.Interplex.ConnectionString = "localhost:1"
	require.NoError(t, AddRemoteCache(cp))

	// read back via GetCacheConfiguration
	c, err := GetCacheConfiguration()
	require.NoError(t, err)
	require.IsType(t, &InterplexCache{}, c)

	// remove remote cache
	require.NoError(t, RemoveRemoteCache())
	// now default should be file-based
	c2, err := GetCacheConfiguration()
	require.NoError(t, err)
	require.IsType(t, &FileBasedCache{}, c2)
}
