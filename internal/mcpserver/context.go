package mcpserver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/martinsuchenak/go-scaffolder/internal/config"
)

func GenerateContext(cfg *config.ProjectConfig, projectDir string) string {
	var sb strings.Builder

	sb.WriteString("# Project Context\n\n")
	fmt.Fprintf(&sb, "- **App Name**: %s\n", cfg.AppName)
	fmt.Fprintf(&sb, "- **Module Path**: %s\n", cfg.ModulePath)
	fmt.Fprintf(&sb, "- **Output Directory**: %s\n", cfg.OutputDir)
	sb.WriteString("\n## Enabled Features\n\n")

	features := []struct {
		Name    string
		Enabled bool
	}{
		{"CLI", cfg.Features.CLI},
		{"API", cfg.Features.API},
		{"MCP", cfg.Features.MCP},
		{"UI", cfg.Features.UI},
		{"DB", cfg.Features.DB},
		{"Cache", cfg.Features.Cache},
		{"Docker", cfg.Features.Docker},
		{"Nomad", cfg.Features.Nomad},
		{"CI", cfg.Features.CI},
		{"Otel", cfg.Features.Otel},
	}

	for _, f := range features {
		status := "disabled"
		if f.Enabled {
			status = "enabled"
		}
		fmt.Fprintf(&sb, "- %s: %s\n", f.Name, status)
	}

	if cfg.Features.DB {
		fmt.Fprintf(&sb, "\n**Database Type**: %s\n", cfg.DBType)
		fmt.Fprintf(&sb, "**Use XDAL**: %t\n", cfg.UseXDAL)
	}
	if cfg.Features.Cache {
		fmt.Fprintf(&sb, "\n**Cache Type**: %s\n", cfg.CacheType)
	}

	sb.WriteString("\n## Available Add Operations\n\n")
	sb.WriteString("- `add_cli_command`: Add a new CLI command\n")
	if cfg.Features.API {
		sb.WriteString("- `add_api_endpoint`: Add a new API endpoint resource (handler, service, storage, routes + tests)\n")
	}
	if cfg.Features.MCP {
		sb.WriteString("- `add_mcp_tool`: Add a new MCP tool\n")
	}

	sb.WriteString("\n## Available Features to Enable\n\n")
	anyAvailable := false
	for _, name := range config.SelectableFeatures {
		if !cfg.Features.HasFeature(name) {
			fmt.Fprintf(&sb, "- `%s`\n", name)
			anyAvailable = true
		}
	}
	if !anyAvailable {
		sb.WriteString("All features are already enabled.\n")
	}

	if projectDir != "" {
		sb.WriteString("\n## Project Structure\n\n")
		sb.WriteString("```\n")
		sb.WriteString(walkProjectTree(projectDir, 0, 3))
		sb.WriteString("```\n")
	}

	return sb.String()
}

func walkProjectTree(dir string, depth, maxDepth int) string {
	if depth > maxDepth {
		return ""
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	var sb strings.Builder
	indent := strings.Repeat("  ", depth)

	skipDirs := map[string]bool{
		".git": true, "node_modules": true, "vendor": true,
		".idea": true, ".vscode": true, "dist": true, "tmp": true,
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") && name != ".go-scaffolder.yaml" {
			continue
		}
		if skipDirs[name] {
			continue
		}

		if entry.IsDir() {
			fmt.Fprintf(&sb, "%s%s/\n", indent, name)
			sb.WriteString(walkProjectTree(filepath.Join(dir, name), depth+1, maxDepth))
		} else {
			fmt.Fprintf(&sb, "%s%s\n", indent, name)
		}
	}

	return sb.String()
}
