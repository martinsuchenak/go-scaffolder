//go:build integration

package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/martinsuchenak/go-scaffolder/internal/config"
	"github.com/martinsuchenak/go-scaffolder/internal/engine"
	"github.com/martinsuchenak/go-scaffolder/internal/writer"
	"github.com/martinsuchenak/go-scaffolder/templates"
)

func scaffoldProject(t *testing.T, cfg *config.ProjectConfig) string {
	t.Helper()
	config.ResolveFeatures(&cfg.Features)
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validation error: %v", err)
	}

	eng := engine.New(templates.FS)
	files, err := eng.RenderAll(cfg)
	if err != nil {
		t.Fatalf("RenderAll error: %v", err)
	}

	outputDir := t.TempDir()
	if err := writer.WriteAll(outputDir, files); err != nil {
		t.Fatalf("WriteAll error: %v", err)
	}

	return outputDir
}

func runInDir(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\nOutput:\n%s", name, args, err, output)
	}
	return string(output)
}

func TestIntegration_APIandDB(t *testing.T) {
	cfg := &config.ProjectConfig{
		AppName:    "test-api-db",
		OutputDir:  "/tmp/test-api-db",
		ModulePath: "test-api-db",
		Features: config.FeatureSet{
			API:    true,
			DB:     true,
			Docker: true,
		},
		DBType: config.DBPostgreSQL,
	}

	dir := scaffoldProject(t, cfg)
	runInDir(t, dir, "go", "mod", "tidy")
	runInDir(t, dir, "go", "build", "./...")
	runInDir(t, dir, "go", "vet", "./...")
	runInDir(t, dir, "go", "test", "./...")
}

func TestIntegration_AllFeatures(t *testing.T) {
	cfg := &config.ProjectConfig{
		AppName:    "test-all-features",
		OutputDir:  "/tmp/test-all",
		ModulePath: "test-all-features",
		Features: config.FeatureSet{
			API:    true,
			MCP:    true,
			UI:     true,
			DB:     true,
			Cache:  true,
			Docker: true,
			Nomad:  true,
		},
		DBType:    config.DBPostgreSQL,
		CacheType: config.CacheRedis,
	}

	dir := scaffoldProject(t, cfg)
	runInDir(t, dir, "go", "mod", "tidy")
	runInDir(t, dir, "go", "build", "./...")
	runInDir(t, dir, "go", "vet", "./...")
	runInDir(t, dir, "go", "test", "./...")
}

func TestIntegration_MinimalCLI(t *testing.T) {
	cfg := &config.ProjectConfig{
		AppName:    "test-minimal",
		OutputDir:  "/tmp/test-minimal",
		ModulePath: "test-minimal",
		Features:   config.FeatureSet{},
	}

	dir := scaffoldProject(t, cfg)
	runInDir(t, dir, "go", "mod", "tidy")
	runInDir(t, dir, "go", "build", "./...")
	runInDir(t, dir, "go", "vet", "./...")
	runInDir(t, dir, "go", "test", "./...")
}

func TestIntegration_ConfigFileParity(t *testing.T) {
	cfg1 := &config.ProjectConfig{
		AppName:    "test-parity",
		OutputDir:  "/tmp/test-parity",
		ModulePath: "test-parity",
		Features: config.FeatureSet{
			API:    true,
			Docker: true,
		},
	}

	cfg2 := &config.ProjectConfig{
		AppName:    "test-parity",
		OutputDir:  "/tmp/test-parity",
		ModulePath: "test-parity",
		Features: config.FeatureSet{
			API:    true,
			Docker: true,
		},
	}

	config.ResolveFeatures(&cfg1.Features)
	config.ResolveFeatures(&cfg2.Features)

	eng := engine.New(templates.FS)
	files1, err := eng.RenderAll(cfg1)
	if err != nil {
		t.Fatalf("RenderAll (1) error: %v", err)
	}
	files2, err := eng.RenderAll(cfg2)
	if err != nil {
		t.Fatalf("RenderAll (2) error: %v", err)
	}

	if len(files1) != len(files2) {
		t.Fatalf("file count mismatch: %d vs %d", len(files1), len(files2))
	}

	for path, content1 := range files1 {
		content2, ok := files2[path]
		if !ok {
			t.Fatalf("file %q missing from second render", path)
		}
		if string(content1) != string(content2) {
			t.Fatalf("file %q content mismatch", path)
		}
	}
}

func TestIntegration_CacheSQLite(t *testing.T) {
	cfg := &config.ProjectConfig{
		AppName:    "test-cache-sqlite",
		OutputDir:  "/tmp/test-cache-sqlite",
		ModulePath: "test-cache-sqlite",
		Features: config.FeatureSet{
			DB:    true,
			Cache: true,
		},
		DBType:    config.DBSQLite,
		CacheType: config.CacheValkey,
	}

	dir := scaffoldProject(t, cfg)
	runInDir(t, dir, "go", "mod", "tidy")
	runInDir(t, dir, "go", "build", "./...")
	runInDir(t, dir, "go", "vet", "./...")
	runInDir(t, dir, "go", "test", "./...")
}

func TestIntegration_MySQLRedis(t *testing.T) {
	cfg := &config.ProjectConfig{
		AppName:    "test-mysql-redis",
		OutputDir:  "/tmp/test-mysql-redis",
		ModulePath: "test-mysql-redis",
		Features: config.FeatureSet{
			API:   true,
			DB:    true,
			Cache: true,
		},
		DBType:    config.DBMySQL,
		CacheType: config.CacheRedis,
	}

	dir := scaffoldProject(t, cfg)
	runInDir(t, dir, "go", "mod", "tidy")
	runInDir(t, dir, "go", "build", "./...")
	runInDir(t, dir, "go", "vet", "./...")
	runInDir(t, dir, "go", "test", "./...")
}

func TestIntegration_ConfigFileScaffolding(t *testing.T) {
	dir := t.TempDir()
	configYaml := `app_name: test-config-file
output_dir: ` + filepath.Join(dir, "output") + `
features:
  - api
  - db
  - docker
db_type: postgresql
`
	configPath := filepath.Join(dir, "scaffold.yaml")
	if err := os.WriteFile(configPath, []byte(configYaml), 0644); err != nil {
		t.Fatal(err)
	}

	binary := filepath.Join(dir, "scaffolder")
	buildCmd := exec.Command("go", "build", "-o", binary, ".")
	buildCmd.Dir = "/Users/martinsuchenak/Devel/projects/new/go-scaffolder"
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build error: %v\n%s", err, output)
	}

	scaffoldCmd := exec.Command(binary, "--config="+configPath)
	scaffoldCmd.Dir = dir
	if output, err := scaffoldCmd.CombinedOutput(); err != nil {
		t.Fatalf("scaffold error: %v\n%s", err, output)
	}

	outputDir := filepath.Join(dir, "output")
	if _, err := os.Stat(filepath.Join(outputDir, "main.go")); os.IsNotExist(err) {
		t.Fatal("main.go should exist in output directory")
	}
	if _, err := os.Stat(filepath.Join(outputDir, "go.mod")); os.IsNotExist(err) {
		t.Fatal("go.mod should exist in output directory")
	}
}
