package engine

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ExternalManifest struct {
	Templates []ExternalTemplateEntry `yaml:"templates"`
}

type ExternalTemplateEntry struct {
	TemplatePath     string   `yaml:"template"`
	OutputPath       string   `yaml:"output"`
	RequiredFeatures []string `yaml:"features,omitempty"`
}

func LoadExternalTemplates(dir string) (fs.FS, []TemplateEntry, error) {
	manifestPath := filepath.Join(dir, "manifest.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, nil, fmt.Errorf("reading manifest.yaml: %w", err)
	}

	var manifest ExternalManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, nil, fmt.Errorf("parsing manifest.yaml: %w", err)
	}

	entries := make([]TemplateEntry, 0, len(manifest.Templates))
	for _, ext := range manifest.Templates {
		if ext.TemplatePath == "" || ext.OutputPath == "" {
			return nil, nil, fmt.Errorf("manifest entry missing template or output path")
		}

		fullPath := filepath.Join(dir, ext.TemplatePath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("template file not found: %s", ext.TemplatePath)
		}

		entries = append(entries, TemplateEntry{
			TemplatePath:     ext.TemplatePath,
			OutputPath:       ext.OutputPath,
			RequiredFeatures: ext.RequiredFeatures,
		})
	}

	externalFS := os.DirFS(dir)
	return externalFS, entries, nil
}
