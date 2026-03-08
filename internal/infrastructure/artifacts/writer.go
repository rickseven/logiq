package artifacts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteArtifact serializes data to JSON and saves it in the .logiq/artifacts directory.
func WriteArtifact(name string, data interface{}) error {
	dir := filepath.Join(".logiq", "artifacts")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create artifacts directory: %w", err)
	}

	path := filepath.Join(dir, name)
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create artifact file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode json: %w", err)
	}

	return nil
}
