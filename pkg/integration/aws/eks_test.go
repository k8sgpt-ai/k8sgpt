package aws

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetKubeconfigPath(t *testing.T) {
	t.Run("uses explicit kubeconfig from viper when set", func(t *testing.T) {
		viper.Reset()
		t.Cleanup(viper.Reset)
		explicit := filepath.Join("custom", "kubeconfig")
		viper.Set("kubeconfig", explicit)

		got, err := getKubeconfigPath()

		require.NoError(t, err)
		assert.Equal(t, explicit, got)
	})

	t.Run("falls back to the user home dir kubeconfig", func(t *testing.T) {
		viper.Reset()
		t.Cleanup(viper.Reset)

		home, err := os.UserHomeDir()
		require.NoError(t, err)

		got, err := getKubeconfigPath()

		require.NoError(t, err)
		assert.Equal(t, filepath.Join(home, ".kube", "config"), got)
		// The fallback must resolve against the user's home directory rather
		// than collapsing to a relative ".kube/config", which is what happened
		// on Windows where the HOME environment variable is typically unset.
		assert.True(t, filepath.IsAbs(got))
	})
}
