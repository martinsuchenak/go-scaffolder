package writer

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteAll writes files under outputDir. Paths must be local (relative, no
// ".." escapes) and writes go through an os.Root, so neither a crafted path
// nor a symlink inside outputDir can place a file outside it.
func WriteAll(outputDir string, files map[string][]byte) error {
	root, err := os.OpenRoot(outputDir)
	if err != nil {
		return fmt.Errorf("opening output directory %s: %w", outputDir, err)
	}
	defer func() { _ = root.Close() }()

	for relPath, content := range files {
		if !filepath.IsLocal(relPath) {
			return fmt.Errorf("refusing to write %q: path escapes output directory", relPath)
		}

		dir := filepath.Dir(relPath)
		if err := root.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", filepath.Join(outputDir, dir), err)
		}

		if err := root.WriteFile(relPath, content, 0644); err != nil {
			return fmt.Errorf("writing file %s: %w", filepath.Join(outputDir, relPath), err)
		}
	}
	return nil
}
