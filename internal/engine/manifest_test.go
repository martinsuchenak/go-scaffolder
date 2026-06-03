package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/martinsuchenak/go-scaffolder/internal/config"
	"github.com/martinsuchenak/go-scaffolder/templates"
)

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadExternalTemplates(t *testing.T) {
	dir := t.TempDir()

	manifest := `templates:
  - template: custom/hello.go.tmpl
    output: internal/custom/hello.go
    features:
      - myfeature
  - template: extra.txt.tmpl
    output: extra.txt
`
	if err := os.WriteFile(filepath.Join(dir, "manifest.yaml"), []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "custom"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "custom/hello.go.tmpl"), []byte("package custom\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "extra.txt.tmpl"), []byte("extra content\n"), 0644); err != nil {
		t.Fatal(err)
	}

	extFS, entries, err := LoadExternalTemplates(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if extFS == nil {
		t.Fatal("expected non-nil FS")
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].OutputPath != "internal/custom/hello.go" {
		t.Errorf("unexpected output path: %s", entries[0].OutputPath)
	}
	if len(entries[0].RequiredFeatures) != 1 || entries[0].RequiredFeatures[0] != "myfeature" {
		t.Errorf("unexpected features: %v", entries[0].RequiredFeatures)
	}
}

func TestLoadExternalTemplatesMissingManifest(t *testing.T) {
	dir := t.TempDir()
	_, _, err := LoadExternalTemplates(dir)
	if err == nil {
		t.Fatal("expected error for missing manifest")
	}
}

func TestLoadExternalTemplatesMissingFile(t *testing.T) {
	dir := t.TempDir()
	manifest := `templates:
  - template: nonexistent.tmpl
    output: out.txt
`
	writeFile(t, filepath.Join(dir, "manifest.yaml"), []byte(manifest))

	_, _, err := LoadExternalTemplates(dir)
	if err == nil {
		t.Fatal("expected error for missing template file")
	}
}

func TestExternalTemplateOverride(t *testing.T) {
	dir := t.TempDir()

	// Override the built-in Taskfile.yml template
	manifest := `templates:
  - template: Taskfile.yml.tmpl
    output: Taskfile.yml
`
	writeFile(t, filepath.Join(dir, "manifest.yaml"), []byte(manifest))
	writeFile(t, filepath.Join(dir, "Taskfile.yml.tmpl"), []byte("# Custom Taskfile for {{.AppName}}\n"))

	extFS, entries, err := LoadExternalTemplates(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	eng := New(templates.FS)
	eng.AddExternalTemplates(extFS, entries)

	cfg := &config.ProjectConfig{
		AppName:    "test-override",
		OutputDir:  "/tmp/test",
		ModulePath: "test-override",
		Features:   config.FeatureSet{CLI: true},
	}

	files, err := eng.RenderAll(cfg)
	if err != nil {
		t.Fatalf("RenderAll error: %v", err)
	}

	taskfile := string(files["Taskfile.yml"])
	if !strings.Contains(taskfile, "# Custom Taskfile") {
		t.Fatal("expected external template to override built-in Taskfile.yml")
	}
	if strings.Contains(taskfile, "CGO_ENABLED") {
		t.Fatal("built-in Taskfile content should not be present")
	}
}

func TestExternalTemplateCustomFeature(t *testing.T) {
	dir := t.TempDir()

	manifest := `templates:
  - template: custom.go.tmpl
    output: internal/custom/custom.go
    features:
      - my_module
`
	writeFile(t, filepath.Join(dir, "manifest.yaml"), []byte(manifest))
	writeFile(t, filepath.Join(dir, "custom.go.tmpl"), []byte("package custom\n// {{.AppName}}\n"))

	extFS, entries, err := LoadExternalTemplates(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	eng := New(templates.FS)
	eng.AddExternalTemplates(extFS, entries)

	// Without the custom tag -- file should NOT be included
	cfg := &config.ProjectConfig{
		AppName:    "test-custom",
		OutputDir:  "/tmp/test",
		ModulePath: "test-custom",
		Features:   config.FeatureSet{CLI: true},
	}

	files, err := eng.RenderAll(cfg)
	if err != nil {
		t.Fatalf("RenderAll error: %v", err)
	}
	if _, ok := files["internal/custom/custom.go"]; ok {
		t.Fatal("custom template should NOT be included without the custom tag")
	}

	// With the custom tag -- file SHOULD be included
	cfg.CustomTags = []string{"my_module"}

	files, err = eng.RenderAll(cfg)
	if err != nil {
		t.Fatalf("RenderAll error: %v", err)
	}
	content, ok := files["internal/custom/custom.go"]
	if !ok {
		t.Fatal("custom template should be included with the custom tag")
	}
	if !strings.Contains(string(content), "test-custom") {
		t.Fatal("custom template should contain the AppName")
	}
}

func TestExternalTemplateAddNew(t *testing.T) {
	dir := t.TempDir()

	manifest := `templates:
  - template: monitoring.go.tmpl
    output: internal/monitoring/monitoring.go
`
	writeFile(t, filepath.Join(dir, "manifest.yaml"), []byte(manifest))
	writeFile(t, filepath.Join(dir, "monitoring.go.tmpl"), []byte("package monitoring\n"))

	extFS, entries, err := LoadExternalTemplates(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	eng := New(templates.FS)
	eng.AddExternalTemplates(extFS, entries)

	cfg := &config.ProjectConfig{
		AppName:    "test-add",
		OutputDir:  "/tmp/test",
		ModulePath: "test-add",
		Features:   config.FeatureSet{CLI: true},
	}

	files, err := eng.RenderAll(cfg)
	if err != nil {
		t.Fatalf("RenderAll error: %v", err)
	}

	if _, ok := files["internal/monitoring/monitoring.go"]; !ok {
		t.Fatal("new external template should be included")
	}
	// Built-in templates should still be present
	if _, ok := files["main.go"]; !ok {
		t.Fatal("built-in main.go should still be present")
	}
}

func TestLoadExternalTemplatesInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "manifest.yaml"), []byte("{{invalid yaml"))
	_, _, err := LoadExternalTemplates(dir)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadExternalTemplatesEmptyPaths(t *testing.T) {
	dir := t.TempDir()
	manifest := `templates:
  - template: ""
    output: out.txt`
	writeFile(t, filepath.Join(dir, "manifest.yaml"), []byte(manifest))
	_, _, err := LoadExternalTemplates(dir)
	if err == nil {
		t.Fatal("expected error for empty template path")
	}

	manifest2 := `templates:
  - template: something.tmpl
    output: ""`
	writeFile(t, filepath.Join(dir, "manifest.yaml"), []byte(manifest2))
	writeFile(t, filepath.Join(dir, "something.tmpl"), []byte("content"))
	_, _, err = LoadExternalTemplates(dir)
	if err == nil {
		t.Fatal("expected error for empty output path")
	}
}

func TestMultipleExternalTemplateSources(t *testing.T) {
	dir1 := t.TempDir()
	manifest1 := `templates:
  - template: Taskfile.yml.tmpl
    output: Taskfile.yml`
	writeFile(t, filepath.Join(dir1, "manifest.yaml"), []byte(manifest1))
	writeFile(t, filepath.Join(dir1, "Taskfile.yml.tmpl"), []byte("# Custom from source 1\n"))

	dir2 := t.TempDir()
	manifest2 := `templates:
  - template: main.go.tmpl
    output: main.go`
	writeFile(t, filepath.Join(dir2, "manifest.yaml"), []byte(manifest2))
	writeFile(t, filepath.Join(dir2, "main.go.tmpl"), []byte("// Custom from source 2\n"))

	extFS1, entries1, err := LoadExternalTemplates(dir1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	extFS2, entries2, err := LoadExternalTemplates(dir2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	eng := New(templates.FS)
	eng.AddExternalTemplates(extFS1, entries1)
	eng.AddExternalTemplates(extFS2, entries2)

	cfg := &config.ProjectConfig{
		AppName:    "test-multi",
		OutputDir:  "/tmp/test",
		ModulePath: "test-multi",
		Features:   config.FeatureSet{CLI: true},
	}

	files, err := eng.RenderAll(cfg)
	if err != nil {
		t.Fatalf("RenderAll error: %v", err)
	}

	taskfile := string(files["Taskfile.yml"])
	if !strings.Contains(taskfile, "# Custom from source 1") {
		t.Fatal("expected override from source 1")
	}

	mainGo := string(files["main.go"])
	if !strings.Contains(mainGo, "// Custom from source 2") {
		t.Fatal("expected override from source 2")
	}
}
