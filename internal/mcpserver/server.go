package mcpserver

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paularlott/mcp"

	"github.com/martinsuchenak/go-scaffolder/internal/config"
	"github.com/martinsuchenak/go-scaffolder/internal/configfile"
	"github.com/martinsuchenak/go-scaffolder/internal/diff"
	"github.com/martinsuchenak/go-scaffolder/internal/engine"
	"github.com/martinsuchenak/go-scaffolder/internal/patcher"
	"github.com/martinsuchenak/go-scaffolder/templates"
)

const projectDirDesc = "Absolute path to the project root directory (where .go-scaffolder.yaml lives). Optional; if omitted the server looks in its working directory."
const stateFileDesc = "Content of .go-scaffolder.yaml as a string. Optional; use for fully remote operation when the server cannot access the project filesystem. Takes precedence over project_dir."
const appNameDesc = "Application name. Optional; used as fallback when no state file is available."
const modulePathDesc = "Go module path. Optional; used as fallback when no state file is available (defaults to app_name)."

func projectParams() []mcp.Parameter {
	return []mcp.Parameter{
		mcp.String("project_dir", projectDirDesc),
		mcp.String("state_file", stateFileDesc),
		mcp.String("app_name", appNameDesc),
		mcp.String("module_path", modulePathDesc),
	}
}

func NewServer(version string) *mcp.Server {
	server := mcp.NewServer("go-scaffolder", version)

	server.RegisterTool(
		mcp.NewTool("scaffold", "Scaffold a new Go microservice project. Returns unified diff of all files.",
			mcp.String("app_name", "Application name", mcp.Required()),
			mcp.String("module_path", "Go module path (defaults to app_name)"),
			mcp.String("features", "Comma-separated features: api,mcp,ui,db,cache,docker,nomad"),
			mcp.String("db_type", "Database type: mysql, postgresql, sqlite"),
			mcp.String("cache_type", "Cache type: redis, valkey"),
		),
		handleScaffold,
	)

	addCliParams := append([]mcp.Parameter{
		mcp.String("name", "Name of the new command", mcp.Required()),
	}, projectParams()...)
	server.RegisterTool(
		mcp.NewTool("add_cli_command", "Add a new CLI command to an existing scaffolded project. Returns unified diff.", addCliParams...),
		func(ctx context.Context, req *mcp.ToolRequest) (*mcp.ToolResponse, error) {
			return handleAdd(ctx, req, "cli-command", "")
		},
	)

	addApiParams := append([]mcp.Parameter{
		mcp.String("name", "Name of the new endpoint resource", mcp.Required()),
	}, projectParams()...)
	server.RegisterTool(
		mcp.NewTool("add_api_endpoint", "Add a new API endpoint resource to an existing scaffolded project. Returns unified diff.", addApiParams...),
		func(ctx context.Context, req *mcp.ToolRequest) (*mcp.ToolResponse, error) {
			return handleAdd(ctx, req, "api-endpoint", "api")
		},
	)

	addMcpParams := append([]mcp.Parameter{
		mcp.String("name", "Name of the new MCP tool", mcp.Required()),
	}, projectParams()...)
	server.RegisterTool(
		mcp.NewTool("add_mcp_tool", "Add a new MCP tool to an existing scaffolded project. Returns unified diff.", addMcpParams...),
		func(ctx context.Context, req *mcp.ToolRequest) (*mcp.ToolResponse, error) {
			return handleAdd(ctx, req, "mcp-tool", "mcp")
		},
	)

	enableParams := append([]mcp.Parameter{
		mcp.String("feature", "Feature to enable: api, mcp, ui, db, cache, docker, nomad", mcp.Required()),
		mcp.String("db_type", "Database type: mysql, postgresql, sqlite (required when enabling db)"),
		mcp.String("cache_type", "Cache type: redis, valkey (required when enabling cache)"),
	}, projectParams()...)
	server.RegisterTool(
		mcp.NewTool("enable_feature", "Enable a feature on an existing scaffolded project. Returns unified diff of new and modified files.", enableParams...),
		handleEnableFeature,
	)

	server.RegisterTool(
		mcp.NewTool("project_context", "Generate a rich context summary of an existing scaffolded project for LLM consumption.", projectParams()...),
		handleProjectContext,
	)

	return server
}

func resolveConfig(req *mcp.ToolRequest) (*config.ProjectConfig, string, error) {
	stateContent := req.StringOr("state_file", "")
	projectDir := req.StringOr("project_dir", "")

	if stateContent != "" {
		cfg, err := configfile.ParseStateContent(stateContent)
		if err != nil {
			return nil, "", fmt.Errorf("parsing state_file content: %w", err)
		}
		return cfg, projectDir, nil
	}

	if projectDir != "" {
		absDir, err := filepath.Abs(projectDir)
		if err != nil {
			return nil, "", fmt.Errorf("resolving project_dir: %w", err)
		}
		stateFilePath := filepath.Join(absDir, configfile.StateFileName)
		cfg, err := configfile.LoadStateFile(stateFilePath)
		if err != nil {
			return nil, absDir, fmt.Errorf("loading state file from %s: %w", stateFilePath, err)
		}
		return cfg, absDir, nil
	}

	cfg, err := configfile.LoadStateFile(configfile.StateFileName)
	if err == nil {
		return cfg, ".", nil
	}

	appName := req.StringOr("app_name", "")
	if appName == "" {
		return nil, "", fmt.Errorf("no state file found; provide project_dir, state_file, or at minimum app_name")
	}

	modulePath := req.StringOr("module_path", appName)
	fallback := &config.ProjectConfig{
		AppName:    appName,
		OutputDir:  ".",
		ModulePath: modulePath,
		Features:   config.FeatureSet{CLI: true},
	}
	return fallback, "", nil
}

func handleScaffold(ctx context.Context, req *mcp.ToolRequest) (*mcp.ToolResponse, error) {
	appName, err := req.String("app_name")
	if err != nil {
		return nil, fmt.Errorf("app_name is required")
	}

	modulePath := req.StringOr("module_path", appName)
	featuresStr := req.StringOr("features", "")
	dbType := req.StringOr("db_type", "")
	cacheType := req.StringOr("cache_type", "")

	fs := config.FeatureSet{}
	for _, f := range strings.Split(featuresStr, ",") {
		switch strings.TrimSpace(strings.ToLower(f)) {
		case "api":
			fs.API = true
		case "mcp":
			fs.MCP = true
		case "ui":
			fs.UI = true
		case "db":
			fs.DB = true
		case "cache":
			fs.Cache = true
		case "docker":
			fs.Docker = true
		case "nomad":
			fs.Nomad = true
		}
	}
	config.ResolveFeatures(&fs)

	cfg := &config.ProjectConfig{
		AppName:    appName,
		OutputDir:  "./" + appName,
		ModulePath: modulePath,
		Features:   fs,
		DBType:     config.DBType(dbType),
		CacheType:  config.CacheType(cacheType),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	eng := engine.New(templates.FS)
	files, err := eng.RenderAll(cfg)
	if err != nil {
		return nil, fmt.Errorf("template rendering error: %w", err)
	}

	var buf bytes.Buffer
	for path, content := range files {
		buf.WriteString(diff.NewFileDiff(path, string(content)))
	}

	return mcp.NewToolResponseText(buf.String()), nil
}

func handleAdd(ctx context.Context, req *mcp.ToolRequest, addType string, requiredFeature string) (*mcp.ToolResponse, error) {
	cfg, _, err := resolveConfig(req)
	if err != nil {
		return nil, err
	}

	if requiredFeature != "" && !cfg.Features.HasFeature(requiredFeature) {
		return nil, fmt.Errorf("this project does not have the %q feature enabled; provide state_file or project_dir with the feature enabled, or use enable_feature first", requiredFeature)
	}

	name, err := req.String("name")
	if err != nil {
		return nil, fmt.Errorf("name is required")
	}

	cfg.ResourceName = name

	eng := engine.New(templates.FS)
	files, err := eng.RenderAdd(cfg, addType)
	if err != nil {
		return nil, fmt.Errorf("template rendering error: %w", err)
	}

	var buf bytes.Buffer
	for path, content := range files {
		buf.WriteString(diff.NewFileDiff(path, string(content)))
	}

	return mcp.NewToolResponseText(buf.String()), nil
}

func handleEnableFeature(ctx context.Context, req *mcp.ToolRequest) (*mcp.ToolResponse, error) {
	cfg, projectDir, err := resolveConfig(req)
	if err != nil {
		return nil, err
	}

	feature, err := req.String("feature")
	if err != nil {
		return nil, fmt.Errorf("feature is required")
	}

	if cfg.Features.HasFeature(feature) {
		return nil, fmt.Errorf("feature %q is already enabled", feature)
	}

	if feature == "db" {
		dbType := req.StringOr("db_type", "")
		if dbType == "" {
			return nil, fmt.Errorf("db_type is required when enabling db")
		}
		cfg.DBType = config.DBType(dbType)
	}

	if feature == "cache" {
		cacheType := req.StringOr("cache_type", "")
		if cacheType == "" {
			return nil, fmt.Errorf("cache_type is required when enabling cache")
		}
		cfg.CacheType = config.CacheType(cacheType)
	}

	switch feature {
	case "api":
		cfg.Features.API = true
	case "mcp":
		cfg.Features.MCP = true
	case "ui":
		cfg.Features.UI = true
	case "db":
		cfg.Features.DB = true
	case "cache":
		cfg.Features.Cache = true
	case "docker":
		cfg.Features.Docker = true
	case "nomad":
		cfg.Features.Nomad = true
		cfg.Features.Docker = true
	default:
		return nil, fmt.Errorf("unknown feature: %s", feature)
	}

	eng := engine.New(templates.FS)
	var buf bytes.Buffer

	var allEntries []engine.TemplateEntry
	entries := eng.EnableFeatureManifest(feature)
	if entries != nil {
		allEntries = append(allEntries, entries...)
	}
	if feature == "cache" {
		allEntries = append(allEntries, eng.EnableFeatureCacheManifest(cfg.CacheType)...)
	}
	if cfg.Features.NeedsSRVResolve() {
		resolvePath := "internal/resolve/resolve.go"
		if projectDir != "" {
			resolvePath = filepath.Join(projectDir, resolvePath)
		}
		if _, statErr := os.Stat(resolvePath); os.IsNotExist(statErr) {
			allEntries = append(allEntries, eng.EnableFeatureSRVManifest()...)
		}
	}

	if len(allEntries) > 0 {
		files, renderErr := eng.RenderFeatureFiles(cfg, allEntries)
		if renderErr != nil {
			return nil, fmt.Errorf("template rendering error: %w", renderErr)
		}
		for path, content := range files {
			if projectDir != "" {
				fullPath := filepath.Join(projectDir, path)
				if _, statErr := os.Stat(fullPath); statErr == nil {
					continue
				}
			}
			buf.WriteString(diff.NewFileDiff(path, string(content)))
		}
	}

	patches := patcher.FeaturePatches(feature, cfg)
	if len(patches) > 0 {
		if projectDir != "" {
			computed := patcher.ComputePatches(projectDir, patches)
			for _, cp := range computed {
				if cp.Applied {
					buf.WriteString(diff.UnifiedDiff(cp.File, cp.File, cp.OldContent, cp.NewContent))
				} else {
					buf.WriteString(fmt.Sprintf("# Could not compute patch for %s (%s)\n# Add manually:\n%s\n\n", cp.File, cp.Description, cp.Content))
				}
			}
		} else {
			for _, p := range patches {
				content := p.Content
				if p.Replace != nil {
					content = p.Replace.Content
				}
				marker := p.Marker
				if p.Replace != nil {
					marker = p.Replace.StartMarker
				}
				buf.WriteString(fmt.Sprintf("# Patch for %s (%s)\n# Apply near marker: %s\n%s\n\n", p.File, p.Description, marker, content))
			}
		}
	}

	return mcp.NewToolResponseText(buf.String()), nil
}

func handleProjectContext(ctx context.Context, req *mcp.ToolRequest) (*mcp.ToolResponse, error) {
	cfg, projectDir, err := resolveConfig(req)
	if err != nil {
		return nil, err
	}

	result := GenerateContext(cfg, projectDir)
	return mcp.NewToolResponseText(result), nil
}
