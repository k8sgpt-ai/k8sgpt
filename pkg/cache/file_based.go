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

// cachePath resolves key to a file inside the k8sgpt cache directory. The key
// is checked before xdg.CacheFile runs, because xdg.CacheFile creates the
// parent directories of whatever path it is given.
func cachePath(key string) (string, error) {
	if !filepath.IsLocal(key) {
		return "", fmt.Errorf("invalid cache key %q: must be a relative path inside the cache directory", key)
	}
	return xdg.CacheFile(filepath.Join("k8sgpt", key))
}

func (f *FileBasedCache) Configure(cacheInfo CacheProvider) error {
	return nil
}

func (f *FileBasedCache) IsCacheDisabled() bool {
	return f.noCache
}

func (*FileBasedCache) List() ([]CacheObjectDetails, error) {
	path, err := xdg.CacheFile("k8sgpt")
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var result []CacheObjectDetails
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			return nil, err
		}
		result = append(result, CacheObjectDetails{
			Name:      file.Name(),
			UpdatedAt: info.ModTime(),
		})
	}

	return result, nil
}

func (*FileBasedCache) Exists(key string) bool {
	path, err := cachePath(key)

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
	path, err := cachePath(key)

	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (*FileBasedCache) Remove(key string) error {
	path, err := cachePath(key)

	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	return nil
}

func (*FileBasedCache) Store(key string, data string) error {
	path, err := cachePath(key)

	if err != nil {
		return err
	}

	return os.WriteFile(path, []byte(data), 0600)
}

func (s *FileBasedCache) GetName() string {
	return "file"
}

func (s *FileBasedCache) DisableCache() {
	s.noCache = true
}
