package external

import (
	"os"

	"github.com/goccy/go-json"
)

func loadCache[T any](path string, value *T) (found bool, err error) {
	// cache not found
	if _, err := os.Stat(path); err != nil {
		return false, nil
	}

	cache, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	if err = json.Unmarshal(cache, value); err != nil {
		return false, err
	}

	return true, nil
}

func saveCache[T any](path string, result *T) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
