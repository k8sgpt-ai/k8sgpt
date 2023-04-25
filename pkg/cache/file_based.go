package cache

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
)

var _ (ICache) = (*FileBasedCache)(nil)

type FileBasedCache struct{}

func (*FileBasedCache) Exists(key string) bool {
	path, err := xdg.CacheFile(filepath.Join("k8sgpt", key))

	if err != nil {
		fmt.Fprintln(os.Stderr, "warning: error while testing if cache key exists:", err)
		return false
	}

	exists, err := util.FileExists(path)

	if err != nil {
		fmt.Fprintln(os.Stderr, "warning: error while testing if cache key exists:", err)
		return false
	}

	return exists
}

func (*FileBasedCache) Load(key string) (string, error) {
	path, err := xdg.CacheFile(filepath.Join("k8sgpt", key))

	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (*FileBasedCache) Store(key string, data string) error {
	path, err := xdg.CacheFile(filepath.Join("k8sgpt", key))

	if err != nil {
		return err
	}

	return os.WriteFile(path, []byte(data), 0600)
}
