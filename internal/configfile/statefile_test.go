package configfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/martinsuchenak/go-scaffolder/internal/config"
)

func TestWriteAndLoadStateFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, StateFileName)

	cfg := &config.ProjectConfig{
		AppName:    "test-app",
		OutputDir:  "/tmp/test",
		ModulePath: "github.com/test/test-app",
		Features: config.FeatureSet{
			CLI:    true,
			API:    true,
			DB:     true,
			Docker: true,
		},
		DBType: config.DBPostgreSQL,
	}

	if err := WriteStateFile(path, cfg); err != nil {
		t.Fatalf("WriteStateFile error: %v", err)
	}

	loaded, err := LoadStateFile(path)
	if err != nil {
		t.Fatalf("LoadStateFile error: %v", err)
	}

	if loaded.AppName != cfg.AppName {
		t.Errorf("AppName: got %q, want %q", loaded.AppName, cfg.AppName)
	}
	if loaded.ModulePath != cfg.ModulePath {
		t.Errorf("ModulePath: got %q, want %q", loaded.ModulePath, cfg.ModulePath)
	}
	if !loaded.Features.API {
		t.Error("API feature should be enabled")
	}
	if !loaded.Features.DB {
		t.Error("DB feature should be enabled")
	}
	if !loaded.Features.Docker {
		t.Error("Docker feature should be enabled")
	}
	if loaded.DBType != config.DBPostgreSQL {
		t.Errorf("DBType: got %q, want %q", loaded.DBType, config.DBPostgreSQL)
	}
	if loaded.OutputDir != "." {
		t.Errorf("OutputDir: got %q, want %q", loaded.OutputDir, ".")
	}
}

func TestLoadStateFileNotFound(t *testing.T) {
	_, err := LoadStateFile("/nonexistent/.go-scaffolder.yaml")
	if err == nil {
		t.Fatal("expected error for missing state file")
	}
}

func TestWriteStateFileAllFeatures(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, StateFileName)

	cfg := &config.ProjectConfig{
		AppName:    "full-app",
		OutputDir:  "/tmp/full",
		ModulePath: "full-app",
		Features: config.FeatureSet{
			CLI:    true,
			API:    true,
			MCP:    true,
			UI:     true,
			DB:     true,
			Cache:  true,
			Docker: true,
			Nomad:  true,
		},
		DBType:    config.DBMySQL,
		CacheType: config.CacheRedis,
	}

	if err := WriteStateFile(path, cfg); err != nil {
		t.Fatalf("WriteStateFile error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading state file: %v", err)
	}

	content := string(data)
	for _, feature := range []string{"api", "mcp", "ui", "db", "cache", "docker", "nomad"} {
		if !contains(content, feature) {
			t.Errorf("state file should contain feature %q", feature)
		}
	}
}

func TestParseStateContent(t *testing.T) {
	content := `app_name: test-app
module_path: github.com/test/test-app
features:
  - api
  - db
db_type: postgresql
`
	cfg, err := ParseStateContent(content)
	if err != nil {
		t.Fatalf("ParseStateContent error: %v", err)
	}

	if cfg.AppName != "test-app" {
		t.Errorf("AppName: got %q, want %q", cfg.AppName, "test-app")
	}
	if cfg.ModulePath != "github.com/test/test-app" {
		t.Errorf("ModulePath: got %q, want %q", cfg.ModulePath, "github.com/test/test-app")
	}
	if !cfg.Features.API {
		t.Error("API feature should be enabled")
	}
	if !cfg.Features.DB {
		t.Error("DB feature should be enabled")
	}
	if cfg.DBType != config.DBPostgreSQL {
		t.Errorf("DBType: got %q, want %q", cfg.DBType, config.DBPostgreSQL)
	}
	if cfg.OutputDir != "." {
		t.Errorf("OutputDir: got %q, want %q", cfg.OutputDir, ".")
	}
}

func TestParseStateContentInvalid(t *testing.T) {
	_, err := ParseStateContent("{{invalid yaml")
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
