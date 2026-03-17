package mcpserver

import (
	"strings"
	"testing"

	"github.com/martinsuchenak/go-scaffolder/internal/config"
)

func TestGenerateContext_BasicOutput(t *testing.T) {
	cfg := &config.ProjectConfig{
		AppName:    "test-app",
		ModulePath: "github.com/test/test-app",
		OutputDir:  "./test-app",
		Features: config.FeatureSet{
			CLI: true,
			API: true,
			MCP: true,
		},
		DBType:    config.DBPostgreSQL,
		CacheType: config.CacheRedis,
	}

	ctx := GenerateContext(cfg, "")

	checks := []string{
		"test-app",
		"github.com/test/test-app",
		"API: enabled",
		"MCP: enabled",
		"DB: disabled",
		"add_cli_command",
		"add_api_endpoint",
		"add_mcp_tool",
	}

	for _, check := range checks {
		if !strings.Contains(ctx, check) {
			t.Errorf("context missing %q", check)
		}
	}
}

func TestGenerateContext_DisabledFeatureNotInAdd(t *testing.T) {
	cfg := &config.ProjectConfig{
		AppName:    "test-app",
		ModulePath: "github.com/test/test-app",
		Features: config.FeatureSet{
			CLI: true,
		},
	}

	ctx := GenerateContext(cfg, "")

	if strings.Contains(ctx, "add_api_endpoint") {
		t.Error("should not list add_api_endpoint when API is disabled")
	}
	if strings.Contains(ctx, "add_mcp_tool") {
		t.Error("should not list add_mcp_tool when MCP is disabled")
	}
}

func TestGenerateContext_AllFeaturesEnabled(t *testing.T) {
	cfg := &config.ProjectConfig{
		AppName:    "test-app",
		ModulePath: "github.com/test/test-app",
		Features: config.FeatureSet{
			CLI: true, API: true, MCP: true, UI: true,
			DB: true, Cache: true, Docker: true, Nomad: true,
		},
		DBType:    config.DBPostgreSQL,
		CacheType: config.CacheRedis,
	}

	ctx := GenerateContext(cfg, "")

	if !strings.Contains(ctx, "All features are already enabled") {
		t.Error("should indicate all features are enabled")
	}
	if !strings.Contains(ctx, "Database Type") {
		t.Error("should show database type")
	}
	if !strings.Contains(ctx, "Cache Type") {
		t.Error("should show cache type")
	}
}

func TestGenerateContext_WithProjectDir(t *testing.T) {
	cfg := &config.ProjectConfig{
		AppName:    "test-app",
		ModulePath: "github.com/test/test-app",
		Features: config.FeatureSet{
			CLI: true,
		},
	}

	ctx := GenerateContext(cfg, t.TempDir())
	if !strings.Contains(ctx, "Project Structure") {
		t.Error("should include project structure when project_dir is set")
	}
}

func TestGenerateContext_NoProjectDir_NoTree(t *testing.T) {
	cfg := &config.ProjectConfig{
		AppName:    "test-app",
		ModulePath: "github.com/test/test-app",
		Features: config.FeatureSet{
			CLI: true,
		},
	}

	ctx := GenerateContext(cfg, "")
	if strings.Contains(ctx, "Project Structure") {
		t.Error("should not include project structure when project_dir is empty")
	}
}
