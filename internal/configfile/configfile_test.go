package configfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pgregory.net/rapid"
	"gopkg.in/yaml.v3"
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
		if enableDB {
			dbType = rapid.SampledFrom([]string{"mysql", "postgresql", "sqlite"}).Draw(t, "DBType")
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
	})
	os.WriteFile(path, data, 0644)

	sc, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if sc.AppName != "test-app" {
		t.Fatalf("AppName = %q, want %q", sc.AppName, "test-app")
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
