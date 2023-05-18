package cache

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/k8sgpt-ai/k8sgpt/pkg/util"
)

var _ (ICache) = (*FileBasedCache)(nil)

type FileBasedCache struct {
	noCache bool
}

func (f *FileBasedCache) IsCacheDisabled() bool {
	return f.noCache
}

func (*FileBasedCache) List() ([]string, error) {
	path, err := xdg.CacheFile("k8sgpt")
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, file := range files {
		result = append(result, file.Name())
	}

	return result, nil
}

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
