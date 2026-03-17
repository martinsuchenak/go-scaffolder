package writer

import (
	"fmt"
	"os"
	"path/filepath"
)

func WriteAll(outputDir string, files map[string][]byte) error {
	for relPath, content := range files {
		fullPath := filepath.Join(outputDir, relPath)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}

		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			return fmt.Errorf("writing file %s: %w", fullPath, err)
		}
	}
	return nil
}
