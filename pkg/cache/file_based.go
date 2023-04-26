package cache

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
)

var _ (ICache) = (*FileBasedCache)(nil)

type FileBasedCache struct {
	noCache bool
}

func (f *FileBasedCache) IsCacheDisabled() bool {
	return f.noCache
}

func (*FileBasedCache) Load(key string) (string, error) {
	path, err := xdg.CacheFile(filepath.Join("k8sgpt", "analysis"))

	if err != nil {
		return "", err
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Créer le fichier s'il n'existe pas
			file, err = os.Create(path)
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		// Faites quelque chose avec la ligne (ex: traitement des données)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == strings.TrimSpace(key) {
			return strings.TrimSpace(parts[1]), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", nil
}

func (*FileBasedCache) Store(key string, data string) error {
	path, err := xdg.CacheFile(filepath.Join("k8sgpt", "analysis"))
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	exists := false
	pattern := fmt.Sprintf("%s:%s", key, data)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, key+":") {
			lines = append(lines, pattern)
			exists = true
		} else {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if !exists {
		lines = append(lines, pattern)
	}

	file.Truncate(0)
	file.Seek(0, 0)

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
}
