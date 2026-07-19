package mcpserver

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paularlott/mcp"

	"github.com/martinsuchenak/go-scaffolder/internal/config"
)

func TestNewServer(t *testing.T) {
	server := NewServer("test", ".")
	if server == nil {
		t.Fatal("NewServer returned nil")
	}
}

func getResponseText(resp *mcp.ToolResponse) string {
	if len(resp.Content) > 0 && resp.Content[0].Type == "text" {
		return resp.Content[0].Text
	}
	return ""
}

func TestHandleScaffold_Basic(t *testing.T) {
	req := mcp.NewToolRequest(map[string]interface{}{
		"app_name":    "test-scaffold",
		"module_path": "github.com/test/scaffold",
	})

	resp, err := handleScaffold(context.Background(), req)
	if err != nil {
		t.Fatalf("handleScaffold error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	content := getResponseText(resp)
	if !strings.Contains(content, "--- /dev/null") {
		t.Error("response should contain diff headers")
	}
	if !strings.Contains(content, "main.go") {
		t.Error("response should include main.go")
	}
}

func TestHandleScaffold_WithFeatures(t *testing.T) {
	req := mcp.NewToolRequest(map[string]interface{}{
		"app_name":    "test-features",
		"module_path": "github.com/test/features",
		"features":    "api,db,docker",
		"db_type":     "postgresql",
	})

	resp, err := handleScaffold(context.Background(), req)
	if err != nil {
		t.Fatalf("handleScaffold error: %v", err)
	}

	content := getResponseText(resp)
	if !strings.Contains(content, "openapi.yaml") {
		t.Error("should include openapi.yaml when API enabled")
	}
	if !strings.Contains(content, "internal/db/db.go") {
		t.Error("should include db.go when DB enabled")
	}
	if !strings.Contains(content, "Dockerfile") {
		t.Error("should include Dockerfile when Docker enabled")
	}
}

func TestHandleScaffold_MissingAppName(t *testing.T) {
	req := mcp.NewToolRequest(map[string]interface{}{})
	_, err := handleScaffold(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing app_name")
	}
}

func TestHandleAdd_UnknownType(t *testing.T) {
	req := mcp.NewToolRequest(map[string]interface{}{
		"app_name": "test",
	})
	_, err := handleAdd(context.Background(), req, "unknown-type", "", ".")
	if err == nil {
		t.Fatal("expected error for unknown add type")
	}
}

func TestHandleAdd_RejectsPathTraversalName(t *testing.T) {
	req := mcp.NewToolRequest(map[string]interface{}{
		"app_name": "test",
		"name":     "../../../../tmp/evil",
	})
	_, err := handleAdd(context.Background(), req, "cli-command", "", ".")
	if err == nil {
		t.Fatal("expected error for path-traversal resource name")
	}
}

func TestResolveConfig_StateFile(t *testing.T) {
	req := mcp.NewToolRequest(map[string]interface{}{
		"state_file": `app_name: test-state
output_dir: /tmp/test
features:
  - api`,
	})

	cfg, _, err := resolveConfig(req, ".")
	if err != nil {
		t.Fatalf("resolveConfig error: %v", err)
	}
	if cfg.AppName != "test-state" {
		t.Errorf("AppName = %q, want test-state", cfg.AppName)
	}
	if !cfg.Features.API {
		t.Error("API feature should be enabled")
	}
}

func TestResolveConfig_ProjectDirEscapeRejected(t *testing.T) {
	req := mcp.NewToolRequest(map[string]interface{}{
		"app_name":    "test",
		"project_dir": "../../../../etc",
	})

	_, _, err := resolveConfig(req, ".")
	if err == nil {
		t.Fatal("expected error for project_dir escaping the server root")
	}
}

func TestResolveConfig_ProjectDirAbsoluteEscapeRejected(t *testing.T) {
	req := mcp.NewToolRequest(map[string]interface{}{
		"app_name":    "test",
		"project_dir": "/etc",
	})

	_, _, err := resolveConfig(req, t.TempDir())
	if err == nil {
		t.Fatal("expected error for absolute project_dir outside the server root")
	}
}

func TestResolveProjectDir_WithinRootAllowed(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "sub")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}

	got, err := resolveProjectDir(root, "sub")
	if err != nil {
		t.Fatalf("resolveProjectDir error: %v", err)
	}
	if got != sub {
		t.Errorf("resolveProjectDir = %q, want %q", got, sub)
	}
}

func TestResolveConfig_NoStateNoApp(t *testing.T) {
	req := mcp.NewToolRequest(map[string]interface{}{})
	_, _, err := resolveConfig(req, ".")
	if err == nil {
		t.Fatal("expected error when no state file and no app_name")
	}
}

func TestResolveConfig_FallbackAppName(t *testing.T) {
	req := mcp.NewToolRequest(map[string]interface{}{
		"app_name": "fallback-app",
	})

	cfg, _, err := resolveConfig(req, ".")
	if err != nil {
		t.Fatalf("resolveConfig error: %v", err)
	}
	if cfg.AppName != "fallback-app" {
		t.Errorf("AppName = %q, want fallback-app", cfg.AppName)
	}
	if !cfg.Features.CLI {
		t.Error("CLI should always be enabled in fallback config")
	}
}

func TestBoolParam(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"", false},
		{"true", true},
		{"false", false},
		{"True", true},
		{"1", true},
		{"0", false},
	}

	for _, tt := range tests {
		got, err := boolParam(tt.input)
		if err != nil {
			t.Errorf("boolParam(%q) error: %v", tt.input, err)
		}
		if got != tt.want {
			t.Errorf("boolParam(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}

	_, err := boolParam("invalid")
	if err == nil {
		t.Error("expected error for invalid boolean value")
	}
}

func TestHandleEnableFeature_AlreadyEnabled(t *testing.T) {
	req := mcp.NewToolRequest(map[string]interface{}{
		"state_file": `app_name: test
output_dir: /tmp/test
features:
  - api`,
		"feature": "api",
	})

	_, err := handleEnableFeature(context.Background(), req, ".")
	if err == nil {
		t.Fatal("expected error for already enabled feature")
	}
	if !strings.Contains(err.Error(), "already enabled") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandleEnableFeature_Unknown(t *testing.T) {
	req := mcp.NewToolRequest(map[string]interface{}{
		"app_name": "test",
		"feature":  "unknown-feature",
	})

	_, err := handleEnableFeature(context.Background(), req, ".")
	if err == nil {
		t.Fatal("expected error for unknown feature")
	}
}

func TestHandleEnableFeature_DB_MissingType(t *testing.T) {
	req := mcp.NewToolRequest(map[string]interface{}{
		"app_name": "test",
		"feature":  "db",
	})

	_, err := handleEnableFeature(context.Background(), req, ".")
	if err == nil {
		t.Fatal("expected error for missing db_type")
	}
}

func TestHandleEnableFeature_Cache_MissingType(t *testing.T) {
	req := mcp.NewToolRequest(map[string]interface{}{
		"app_name": "test",
		"feature":  "cache",
	})

	_, err := handleEnableFeature(context.Background(), req, ".")
	if err == nil {
		t.Fatal("expected error for missing cache_type")
	}
}

func TestHandleProjectContext(t *testing.T) {
	req := mcp.NewToolRequest(map[string]interface{}{
		"state_file": `app_name: test-ctx
output_dir: /tmp/test
features:
  - api`,
	})

	resp, err := handleProjectContext(context.Background(), req, ".")
	if err != nil {
		t.Fatalf("handleProjectContext error: %v", err)
	}

	content := getResponseText(resp)
	if !strings.Contains(content, "test-ctx") {
		t.Error("context should contain app name")
	}
	if !strings.Contains(content, "API: enabled") {
		t.Error("context should show API as enabled")
	}
}

func TestGenerateContext_AllFeatures(t *testing.T) {
	cfg := &config.ProjectConfig{
		AppName:    "full-app",
		ModulePath: "github.com/test/full",
		OutputDir:  "./full",
		Features: config.FeatureSet{
			CLI: true, API: true, MCP: true, UI: true,
			DB: true, Cache: true, Docker: true, Nomad: true,
		},
		DBType:    config.DBPostgreSQL,
		UseXDAL:   true,
		CacheType: config.CacheRedis,
	}

	ctx := GenerateContext(cfg, "")
	if !strings.Contains(ctx, "All features are already enabled") {
		t.Error("should indicate all features enabled")
	}
	if !strings.Contains(ctx, "Database Type") {
		t.Error("should show database type")
	}
	if !strings.Contains(ctx, "Cache Type") {
		t.Error("should show cache type")
	}
}

func TestWalkProjectTree_MaxDepth(t *testing.T) {
	ctx := GenerateContext(&config.ProjectConfig{AppName: "test"}, "")
	if strings.Contains(ctx, "Project Structure") {
		t.Error("should not include project structure when dir is empty")
	}
}
