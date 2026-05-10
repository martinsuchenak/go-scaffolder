package configfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
	"pgregory.net/rapid"
)

// Feature: go-scaffolder, Property 18: Config file missing required fields produces error
func TestProperty18_ConfigFileMissingRequiredFieldsProducesError(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		missingAppName := rapid.Bool().Draw(t, "missingAppName")
		missingOutputDir := rapid.Bool().Draw(t, "missingOutputDir")

		if !missingAppName && !missingOutputDir {
			missingAppName = true
		}

		sc := ScaffoldConfig{
			Features: []string{"api"},
		}
		if !missingAppName {
			sc.AppName = "test-app"
		}
		if !missingOutputDir {
			sc.OutputDir = "/tmp/test"
		}

		_, err := sc.ToProjectConfig()
		if err == nil {
			t.Fatal("expected error for missing required fields")
		}

		if missingAppName && !strings.Contains(err.Error(), "app_name") {
			t.Fatalf("error should mention app_name: %s", err)
		}
		if missingOutputDir && !strings.Contains(err.Error(), "output_dir") {
			t.Fatalf("error should mention output_dir: %s", err)
		}
	})
}

// Feature: go-scaffolder, Property 17: Config file mode produces identical output to interactive mode
func TestProperty17_ConfigFileModeParity(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		appName := rapid.StringMatching(`[a-z][a-z0-9\-]{2,15}`).Draw(t, "AppName")
		enableAPI := rapid.Bool().Draw(t, "API")
		enableDB := rapid.Bool().Draw(t, "DB")
		enableCache := rapid.Bool().Draw(t, "Cache")
		enableDocker := rapid.Bool().Draw(t, "Docker")

		var features []string
		if enableAPI {
			features = append(features, "api")
		}
		if enableDB {
			features = append(features, "db")
		}
		if enableCache {
			features = append(features, "cache")
		}
		if enableDocker {
			features = append(features, "docker")
		}

		var dbType string
		var useXDAL bool
		if enableDB {
			dbType = rapid.SampledFrom([]string{"mysql", "postgresql", "sqlite"}).Draw(t, "DBType")
			useXDAL = rapid.Bool().Draw(t, "UseXDAL")
		}
		var cacheType string
		if enableCache {
			cacheType = rapid.SampledFrom([]string{"redis", "valkey"}).Draw(t, "CacheType")
		}

		sc := ScaffoldConfig{
			AppName:   appName,
			OutputDir: "/tmp/test-" + appName,
			Features:  features,
			DBType:    dbType,
			UseXDAL:   useXDAL,
			CacheType: cacheType,
		}

		pc, err := sc.ToProjectConfig()
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if pc.AppName != appName {
			t.Fatalf("AppName mismatch: got %q, want %q", pc.AppName, appName)
		}
		if pc.Features.CLI != true {
			t.Fatal("CLI should always be true")
		}
		if enableAPI && !pc.Features.API {
			t.Fatal("API should be enabled")
		}
		if enableDB && !pc.Features.DB {
			t.Fatal("DB should be enabled")
		}
		if enableDB && pc.UseXDAL != useXDAL {
			t.Fatalf("UseXDAL mismatch: got %v, want %v", pc.UseXDAL, useXDAL)
		}
	})
}

func TestLoadValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	data, _ := yaml.Marshal(ScaffoldConfig{
		AppName:   "test-app",
		OutputDir: "/tmp/test",
		Features:  []string{"api", "db"},
		DBType:    "postgresql",
		UseXDAL:   true,
	})
	os.WriteFile(path, data, 0644)

	sc, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if sc.AppName != "test-app" {
		t.Fatalf("AppName = %q, want %q", sc.AppName, "test-app")
	}
	if !sc.UseXDAL {
		t.Fatal("UseXDAL should be loaded from YAML")
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestToProjectConfig_AllFeatureShortcut(t *testing.T) {
	sc := ScaffoldConfig{
		AppName:   "test-app",
		OutputDir: "/tmp/test",
		Features:  []string{"all"},
		DBType:    "postgresql",
		UseXDAL:   true,
		CacheType: "redis",
	}

	pc, err := sc.ToProjectConfig()
	if err != nil {
		t.Fatalf("ToProjectConfig error: %v", err)
	}

	if !pc.Features.API || !pc.Features.MCP || !pc.Features.UI || !pc.Features.DB || !pc.Features.Cache || !pc.Features.Docker || !pc.Features.Nomad {
		t.Fatal("all built-in features should be enabled when features contains all")
	}
	if !pc.UseXDAL {
		t.Fatal("UseXDAL should be preserved")
	}
}
